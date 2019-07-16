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
	"sync"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/load/batch"
	"github.com/czcorpus/klogproc/load/sredis"
	"github.com/czcorpus/klogproc/load/tail"
	"github.com/czcorpus/klogproc/save/elastic"
	"github.com/czcorpus/klogproc/save/influx"
	"github.com/oschwald/geoip2-golang"
)

func applyLocation(rec conversion.InputRecord, db *geoip2.Reader, outRec conversion.OutputRecord) {
	ip := rec.GetClientIP()
	if ip != nil {
		city, err := db.City(ip)
		if err != nil {
			log.Printf("Failed to fetch GeoIP data for IP %s: %s", ip.String(), err)

		} else {
			outRec.SetLocation(city.Country.Names["en"], float32(city.Location.Latitude),
				float32(city.Location.Longitude), city.Location.TimeZone)
		}
	}
}

// CNKLogProcessor imports parsed log records represented
// as InputRecord instances
type CNKLogProcessor struct {
	appType        string
	anonymousUsers []int
	geoIPDb        *geoip2.Reader
	chunkSize      int
	currIdx        int
	numNonLoggable int
	logTransformer conversion.LogItemTransformer
}

// ProcItem transforms input log record into an output format.
// In case an unsupported record is encountered, nil is returned.
func (clp *CNKLogProcessor) ProcItem(logRec conversion.InputRecord) conversion.OutputRecord {
	if logRec.AgentIsLoggable() {
		rec, err := clp.logTransformer.Transform(logRec, clp.appType, clp.anonymousUsers)
		if err != nil {
			log.Printf("ERROR: failed to transform item %s", logRec)
			return nil
		}
		applyLocation(logRec, clp.geoIPDb, rec)
		return rec
	}
	clp.numNonLoggable++
	return nil
}

// GetAppType returns a string idenfier unique for a concrete application we
// want to archive logs for (e.g. 'kontext', 'syd', ...)
func (clp *CNKLogProcessor) GetAppType() string {
	return clp.appType
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

func processRedisLogs(conf *Conf, queue *sredis.RedisQueue, processor *CNKLogProcessor, destChans ...chan conversion.OutputRecord) {
	for _, item := range queue.GetItems() {
		rec := processor.ProcItem(item)
		if rec != nil {
			for _, ch := range destChans {
				ch <- rec
			}
		}
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
func processLogs(conf *Conf, action string) {
	geoDb, err := geoip2.Open(conf.GeoIPDbPath)
	if err != nil {
		panic(err)
	}
	defer geoDb.Close()

	channelWriteES := make(chan conversion.OutputRecord, conf.ElasticSearch.PushChunkSize*2)
	channelWriteInflux := make(chan conversion.OutputRecord, conf.InfluxDB.PushChunkSize)
	useChunkMode := true
	var rescue ESImportFailHandler

	go func() {
		// TODO - not applicable for all 'case's
		lt, err := GetLogTransformer(conf.AppType)
		if err != nil {
			panic(err) // TODO
		}
		processor := &CNKLogProcessor{
			geoIPDb:        geoDb,
			chunkSize:      conf.ElasticSearch.PushChunkSize,
			appType:        conf.AppType,
			logTransformer: lt,
		}

		switch action {
		case actionRedis:
			if !conf.UsesRedis() {
				panic("Redis not configured") // TODO
			}
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
			processRedisLogs(conf, redisQueue, processor, channelWriteES, channelWriteInflux)
			close(channelWriteES)
			close(channelWriteInflux)

		case actionBatch:
			// TODO test config
			worklog := batch.NewWorklog(conf.LogFiles.WorklogPath)
			log.Printf("INFO: using worklog %s", conf.LogFiles.WorklogPath)
			defer worklog.Save()
			rescue = worklog
			proc := batch.CreateLogFileProcFunc(processor, channelWriteES, channelWriteInflux)
			proc(&conf.LogFiles, conf.LocalTimezone, worklog.GetLastRecord())
			close(channelWriteES)
			close(channelWriteInflux)

		case actionTail:
			var err error
			useChunkMode = false
			geoDb, err := geoip2.Open(conf.GeoIPDbPath)
			if err != nil {
				log.Fatal("ERROR: ", err)
			}
			lineParsers := make(map[string]batch.LineParser)
			logTransformers := make(map[string]conversion.LogItemTransformer)

			for _, f := range conf.LogTail.Files {
				lineParsers[f.AppType], err = batch.NewLineParser(f.AppType)
				if err != nil {
					log.Fatal("ERROR: Failed to initialize parser: ", err)
				}
				logTransformers[f.AppType], err = GetLogTransformer(f.AppType)
				if err != nil {
					log.Fatal("ERROR: Failed to initialize transformer: ", err)
				}
			}
			tail.Run(
				&conf.LogTail,
				func(rec string, appType string) {
					parsed, err := lineParsers[appType].ParseLine(rec, 0, conf.LocalTimezone)
					if err != nil {
						log.Printf("ERROR: %s", err)
						return
					}
					outRec, err := logTransformers[appType].Transform(parsed, appType, conf.AnonymousUsers)
					if err != nil {
						log.Printf("ERROR: %s", err)
						return
					}
					applyLocation(parsed, geoDb, outRec)
					channelWriteES <- outRec
					channelWriteInflux <- outRec
				},
				func() {
					close(channelWriteES)
					close(channelWriteInflux)
					geoDb.Close()
				},
			)
		}
		log.Printf("INFO: Ignored %d non-loggable items (bots etc.)", processor.numNonLoggable)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	// Elasticsearch bulk writes
	go func() {
		defer wg.Done()
		if conf.HasElasticOut() {
			i := 0
			data := make([][]byte, conf.ElasticSearch.PushChunkSize*2+1)
			failed := make([][]byte, 0, 50)
			var esErr error
			for rec := range channelWriteES {
				jsonData, err := rec.ToJSON()
				jsonMeta := elastic.CNKRecordMeta{
					ID:    rec.GetID(),
					Type:  rec.GetType(),
					Index: conf.ElasticSearch.Index,
				}
				jsonMetaES, err2 := (&elastic.ESCNKRecordMeta{Index: jsonMeta}).ToJSON()

				if err == nil && err2 == nil {
					data[i] = jsonMetaES
					data[i+1] = jsonData
					i += 2

				} else {
					log.Print("ERROR: Failed to encode item ", rec.GetTime())
				}
				if i == conf.ElasticSearch.PushChunkSize*2 || !useChunkMode {
					data[i] = []byte("\n")
					esErr = pushDataToElastic(data[:i+1], &conf.ElasticSearch)
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
			if rescue != nil {
				rescue.RescueFailedChunks(failed)
			}

		} else {
			for range channelWriteES {
			}
		}
	}()

	// InfluxDB batch writes
	go func() {
		defer wg.Done()
		if conf.HasInfluxOut() {
			var err error
			client, err := influx.NewRecordWriter(&conf.InfluxDB)
			if err != nil {
				log.Printf("ERROR: %s", err)
			}
			for rec := range channelWriteInflux {
				client.AddRecord(rec)
			}
			err = client.Finish()
			if err != nil {
				log.Printf("ERROR: %s", err)
			}

		} else {
			for range channelWriteInflux {
			}
		}
	}()

	wg.Wait()
}
