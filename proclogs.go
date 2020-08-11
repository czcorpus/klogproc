// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
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
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/czcorpus/klogproc/config"
	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/ctype"
	"github.com/czcorpus/klogproc/fsop"
	"github.com/czcorpus/klogproc/load/batch"
	"github.com/czcorpus/klogproc/load/sredis"
	"github.com/czcorpus/klogproc/save"
	"github.com/czcorpus/klogproc/save/elastic"
	"github.com/czcorpus/klogproc/save/influx"
	"github.com/czcorpus/klogproc/users"
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

// ClientAnalyzer represents an object which is able to recognize
// bots etc. based on IP and/or user agent.
type ClientAnalyzer interface {
	AgentIsMonitor(rec conversion.InputRecord) bool
	AgentIsBot(rec conversion.InputRecord) bool
	HasBlacklistedIP(rec conversion.InputRecord) bool
}

type ProcessOptions struct {
	worklogReset bool
	dryRun       bool
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
	clientAnalyzer ClientAnalyzer
}

func (clp *CNKLogProcessor) recordIsLoggable(logRec conversion.InputRecord) bool {
	isBlacklisted := false
	if clp.clientAnalyzer.HasBlacklistedIP(logRec) {
		isBlacklisted = true
		log.Printf("INFO: Found blacklisted IP %s", logRec.GetClientIP().String())
	}
	return !clp.clientAnalyzer.AgentIsBot(logRec) && !clp.clientAnalyzer.AgentIsMonitor(logRec) &&
		!isBlacklisted && logRec.IsProcessable()
}

// ProcItem transforms input log record into an output format.
// In case an unsupported record is encountered, nil is returned.
func (clp *CNKLogProcessor) ProcItem(logRec conversion.InputRecord, tzShiftMin int) conversion.OutputRecord {
	if clp.recordIsLoggable(logRec) {
		rec, err := clp.logTransformer.Transform(logRec, clp.appType, tzShiftMin, clp.anonymousUsers)
		if err != nil {
			log.Printf("ERROR: failed to transform item %s: %s", logRec, err)
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

func processRedisLogs(conf *config.Main, queue *sredis.RedisQueue, processor *CNKLogProcessor, destChans ...chan<- conversion.OutputRecord) {
	for _, item := range queue.GetItems() {
		rec := processor.ProcItem(item, conf.LogRedis.TZShift)
		if rec != nil {
			for _, ch := range destChans {
				ch <- rec
			}
		}
	}
}

// ---------

// retryRescuedItems inserts (sychronously) rescued raw bulk
// insert items to ElasticSearch. There is no additial rescue here.
// If something goes wrong we stop (while keeping data in
// rescue queue) and refuse to continue.
func retryRescuedItems(appType string, queue *sredis.RedisQueue, conf *elastic.ConnectionConf) error {
	iterator := queue.GetRescuedChunksIterator()
	chunk := iterator.GetNextChunk()
	log.Printf("INFO: Found %d rescued items. I am going to re-insert them.", len(chunk))
	for len(chunk) > 0 {
		err := elastic.BulkWriteRequest(chunk, appType, conf)
		if err != nil {
			return fmt.Errorf("failed to reuse rescued data chunk: %s", err)
		}
		fixed, err := iterator.RemoveVisitedItems()
		if err != nil {
			return fmt.Errorf("failed to remove rescued & applied data chunk: %s", err)
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
func processLogs(conf *config.Main, action string, options *ProcessOptions) {
	geoDb, err := geoip2.Open(conf.GeoIPDbPath)
	userMap := users.EmptyUserMap()
	confPath := filepath.Join(conf.CustomConfDir, "usermap.json")
	if fsop.IsFile(confPath) {
		userMap, err = users.LoadUserMap(confPath)
		if err != nil {
			log.Fatal("FATAL: ", err)
		}
	}
	defer geoDb.Close()

	var clientTypeDetector ClientAnalyzer
	if conf.BotDefsPath != "" {
		clientTypeDetector, err = ctype.LoadFromResource(conf.BotDefsPath)
		if err != nil {
			log.Fatal("FATAL: ", err)
		}
		log.Printf("INFO: using bot definitions from resource %s", conf.BotDefsPath)

	} else {
		clientTypeDetector = &ctype.LegacyClientTypeAnalyzer{}
		log.Print("WARNING: no bots configuration provided (botDefsPath), using legacy analyzer")
	}

	finishEvent := make(chan bool)
	go func() {
		switch action {
		case actionRedis:
			if !conf.UsesRedis() {
				log.Fatal("FATAL: Redis not configured")
			}
			lt, err := GetLogTransformer(conf.LogRedis.AppType, conf.LogRedis.Version, userMap)
			if err != nil {
				log.Fatal(err)
			}
			processor := &CNKLogProcessor{
				geoIPDb:        geoDb,
				chunkSize:      conf.ElasticSearch.PushChunkSize,
				appType:        conf.LogRedis.AppType,
				logTransformer: lt,
				anonymousUsers: conf.AnonymousUsers,
				clientAnalyzer: clientTypeDetector,
			}
			channelWriteES := make(chan conversion.OutputRecord, conf.ElasticSearch.PushChunkSize*2)
			channelWriteInflux := make(chan conversion.OutputRecord, conf.InfluxDB.PushChunkSize)
			redisQueue, err := sredis.OpenRedisQueue(
				conf.LogRedis.Address,
				conf.LogRedis.Database,
				conf.LogRedis.QueueKey,
				0, // TODO
			)
			if err != nil {
				log.Fatal(err)
			}

			err = retryRescuedItems(conf.LogRedis.AppType, redisQueue, &conf.ElasticSearch)
			if err != nil {
				log.Printf("ERROR: %s", err)
			}

			processRedisLogs(conf, redisQueue, processor, channelWriteES, channelWriteInflux)
			var wg sync.WaitGroup
			wg.Add(2)
			go elastic.RunWriteConsumer(conf.LogRedis.AppType, &conf.ElasticSearch, channelWriteES, &wg, redisQueue)
			go influx.RunWriteConsumer(&conf.InfluxDB, channelWriteInflux, &wg)
			wg.Wait()
			close(channelWriteES)
			close(channelWriteInflux)
			finishEvent <- true
			log.Printf("INFO: Ignored %d non-loggable entries (bots, static files etc.)", processor.numNonLoggable)

		case actionBatch:
			lt, err := GetLogTransformer(conf.LogFiles.AppType, conf.LogFiles.Version, userMap)
			if err != nil {
				log.Fatal(err)
			}
			processor := &CNKLogProcessor{
				geoIPDb:        geoDb,
				chunkSize:      conf.ElasticSearch.PushChunkSize,
				appType:        conf.LogFiles.AppType,
				logTransformer: lt,
				anonymousUsers: conf.AnonymousUsers,
				clientAnalyzer: clientTypeDetector,
			}
			channelWriteES := make(chan conversion.OutputRecord, conf.ElasticSearch.PushChunkSize*2)
			channelWriteInflux := make(chan conversion.OutputRecord, conf.InfluxDB.PushChunkSize)
			worklog := batch.NewWorklog(conf.LogFiles.WorklogPath)
			log.Printf("INFO: using worklog %s", conf.LogFiles.WorklogPath)
			if options.worklogReset {
				log.Printf("truncated worklog %v", worklog)
				err := worklog.Reset()
				if err != nil {
					log.Fatalf("FATAL: unable to initialize worklog: %s", err)
				}
			}
			defer worklog.Save()

			var wg sync.WaitGroup
			wg.Add(2)

			if options.dryRun {
				go save.RunWriteConsumer(channelWriteES, channelWriteInflux, &wg)
				log.Print("WARNING: using dry-run mode, output goes to stdout")

			} else {
				go elastic.RunWriteConsumer(conf.LogFiles.AppType, &conf.ElasticSearch, channelWriteES, &wg, worklog)
				go influx.RunWriteConsumer(&conf.InfluxDB, channelWriteInflux, &wg)
			}
			proc := batch.CreateLogFileProcFunc(processor, channelWriteES, channelWriteInflux)
			proc(&conf.LogFiles, worklog.GetLastRecord())
			close(channelWriteES)
			close(channelWriteInflux)
			wg.Wait()
			finishEvent <- true
			log.Printf("INFO: Ignored %d non-loggable entries (bots, static files etc.)", processor.numNonLoggable)

		case actionTail:
			runTailAction(conf, geoDb, userMap, clientTypeDetector, finishEvent)
		}
	}()
	<-finishEvent

}
