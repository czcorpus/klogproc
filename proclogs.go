// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/czcorpus/klogproc/elastic"
	"github.com/czcorpus/klogproc/fetch"
	"github.com/czcorpus/klogproc/fetch/sfiles"
	"github.com/czcorpus/klogproc/fetch/sredis"
	"github.com/czcorpus/klogproc/record"
	"github.com/oschwald/geoip2-golang"
)

// CNKLogProcessor imports parsed log records represented
// as LogRecord instances
type CNKLogProcessor struct {
	geoIPDb   *geoip2.Reader
	chunk     chan *record.CNKRecord
	chunkSize int
	currIdx   int
}

// ProcItem is a callback function called by log parser
func (clp *CNKLogProcessor) ProcItem(appType string, logRec *fetch.LogRecord) {
	if logRec.AgentIsLoggable() {
		rec := record.New(logRec, appType)
		ip := logRec.GetClientIP()
		if ip != nil {
			city, err := clp.geoIPDb.City(ip)
			if err != nil {
				log.Printf("Failed to fetch GeoIP data for IP %s: %s", ip.String(), err)

			} else {
				rec.GeoIP.IP = ip.String()
				rec.GeoIP.CountryName = city.Country.Names["en"]
				rec.GeoIP.Latitude = float32(city.Location.Latitude)
				rec.GeoIP.Longitude = float32(city.Location.Longitude)
				rec.GeoIP.Location[0] = rec.GeoIP.Longitude
				rec.GeoIP.Location[1] = rec.GeoIP.Latitude
				rec.GeoIP.Timezone = city.Location.TimeZone
			}
		}
		clp.chunk <- rec
	}
}

func pushDataToElastic(data [][]byte, esconf *elastic.SearchConf) error {
	esclient := elastic.NewClient(esconf)
	q := bytes.Join(data, []byte("\n"))
	_, err := esclient.Do("POST", "/_bulk", q)
	if err != nil {
		return fmt.Errorf("Failed to push log chunk: %s", err)
	}
	log.Printf("INFO: Inserted chunk of %d items to ElasticSearch\n", (len(data)-1)/2)
	return nil
}

func processRedisLogs(conf *Conf, queue *sredis.RedisQueue, processor *CNKLogProcessor) {
	for _, item := range queue.GetItems() {
		processor.ProcItem(conf.AppType, item)
	}
}

// ---------

// ESImportFailHandler represents an object able to handle (valid)
// log items we failed to insert to ElasticSearch (typically due
// to inavailability)
type ESImportFailHandler interface {
	RescueFailedChunks(chunk [][]byte) error
}

// retryRescuedItems inserts (sychronously) rescued raw bulk
// insert items to ElasticSearch. There is no additial rescue here.
// If something goes wrong we stop (while keeping data in
// rescue queue) and refuse to continue.
func retryRescuedItems(queue *sredis.RedisQueue, conf *elastic.SearchConf) error {
	iterator := queue.GetRescuedChunksIterator()
	chunk := iterator.GetNextChunk()
	for len(chunk) > 0 {
		err := pushDataToElastic(chunk, conf)
		if err != nil {
			return fmt.Errorf("failed to reuse rescued data chunk")
		}
		fixed, err := iterator.RemoveVisitedItems()
		if err != nil {
			return fmt.Errorf("failed to reuse rescued data chunk")
		}
		log.Printf("INFO: Rescued %d bulk insert rows from the previous failed run(s)", fixed)
		chunk = iterator.GetNextChunk()
	}
	return nil
}

// ----

// ProcessLogs runs through all the logs found in configuration and matching
// some basic properties (it is a query, preferably from a human user etc.).
// The "producer" part of the processing runs in a separate goroutine while
// the main goroutine consumes values via a channel and after each
// n-th (conf.ElasticPushChunkSize) item it stores data to the ElasticSearch
// server.
// Based on config, the function reads either from a Redis list object
// or from a directory of files (in such case it keeps a worklog containing
// last loaded value). In case both locations are configured, Redis has
// precedence.
func processLogs(conf *Conf) {
	geoDb, err := geoip2.Open(conf.GeoIPDbPath)
	if err != nil {
		panic(err)
	}
	defer geoDb.Close()

	chunkChannel := make(chan *record.CNKRecord, conf.ElasticSearch.PushChunkSize*2)
	var rescue ESImportFailHandler

	go func() {
		processor := &CNKLogProcessor{
			geoIPDb:   geoDb,
			chunk:     chunkChannel,
			chunkSize: conf.ElasticSearch.PushChunkSize,
		}

		if conf.UsesRedis() {
			redisQueue, err := sredis.OpenRedisQueue(
				conf.LogRedis.Address,
				conf.LogRedis.Database,
				conf.LogRedis.QueueKey,
				conf.LocalTimezone,
			)
			if err != nil {
				panic(err)
			}
			rescue = redisQueue
			err = retryRescuedItems(redisQueue, &conf.ElasticSearch)
			if err != nil {
				log.Fatalf("ERROR: %s. Please fix the problem before running the program again",
					err)
			}
			processRedisLogs(conf, redisQueue, processor)

		} else {
			worklog := sfiles.NewWorklog(conf.LogFiles.WorklogPath)
			defer worklog.Save()
			rescue = worklog
			sfiles.ProcessFileLogs(&conf.LogFiles, conf.AppType, conf.LocalTimezone,
				worklog.GetLastRecord(), processor)
		}
		close(chunkChannel)
	}()

	i := 0
	data := make([][]byte, conf.ElasticSearch.PushChunkSize*2+1)
	failed := make([][]byte, 0, 50)
	var esErr error
	for rec := range chunkChannel {
		jsonData, err := rec.ToJSON()
		jsonMeta := elastic.CNKRecordMeta{
			ID:    rec.ID,
			Type:  rec.Type,
			Index: conf.ElasticSearch.Index,
		}
		jsonMetaES, err2 := (&elastic.ESCNKRecordMeta{Index: jsonMeta}).ToJSON()

		if err == nil && err2 == nil {
			data[i] = jsonMetaES
			data[i+1] = jsonData
			i += 2

		} else {
			log.Print("ERROR: Failed to encode item ", rec.Datetime)
		}
		if i == conf.ElasticSearch.PushChunkSize*2 {
			data[i] = []byte("\n")
			esErr = pushDataToElastic(data, &conf.ElasticSearch)
			if esErr != nil {
				failed = append(failed, data[:i+1]...)
			}
			i = 0
		}
	}
	if i > 0 {
		data[i] = []byte("\n")
		esErr = pushDataToElastic(data[:i+1], &conf.ElasticSearch)
		if esErr != nil {
			log.Printf("ERROR: %s", esErr)
			failed = append(failed, data[:i+1]...)
		}
	}
	rescue.RescueFailedChunks(failed)
}
