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
	"path/filepath"

	"github.com/rs/zerolog/log"

	"klogproc/botwatch"
	"klogproc/config"
	"klogproc/conversion"
	"klogproc/fsop"
	"klogproc/load/batch"
	"klogproc/logbuffer"
	"klogproc/users"

	"github.com/oschwald/geoip2-golang"
)

func applyLocation(rec conversion.InputRecord, db *geoip2.Reader, outRec conversion.OutputRecord) {
	ip := rec.GetClientIP()
	if len(ip) > 0 {
		city, err := db.City(ip)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch GeoIP data for IP %s.", ip.String())

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
	Add(rec conversion.InputRecord)
	GetBotCandidates() []botwatch.IPStats
	StoreBotCandidates()
	ResetBotCandidates()
	Close()
}

type ProcessOptions struct {
	worklogReset  bool
	dryRun        bool
	analysisOnly  bool
	datetimeRange batch.DatetimeRange
}

// CNKLogProcessor imports parsed log records represented
// as InputRecord instances
type CNKLogProcessor struct {
	appType        string
	appVersion     string
	anonymousUsers []int
	geoIPDb        *geoip2.Reader
	chunkSize      int
	numNonLoggable int
	skipAnalysis   bool
	logTransformer conversion.LogItemTransformer
	clientAnalyzer ClientAnalyzer
	logBuffer      logbuffer.AbstractStorage[conversion.InputRecord]
}

func (clp *CNKLogProcessor) recordIsLoggable(logRec conversion.InputRecord) bool {
	isBlacklisted := false
	if clp.clientAnalyzer.HasBlacklistedIP(logRec) {
		isBlacklisted = true
		log.Info().Msgf("Found blacklisted IP %s", logRec.GetClientIP().String())
	}
	return !clp.clientAnalyzer.AgentIsBot(logRec) && !clp.clientAnalyzer.AgentIsMonitor(logRec) &&
		!isBlacklisted && logRec.IsProcessable()
}

// ProcItem transforms input log record into an output format.
// In case an unsupported record is encountered, nil is returned.
func (clp *CNKLogProcessor) ProcItem(logRec conversion.InputRecord, tzShiftMin int) []conversion.OutputRecord {
	if !clp.skipAnalysis {
		clp.clientAnalyzer.Add(logRec)
	}
	if clp.recordIsLoggable(logRec) {
		ans := make([]conversion.OutputRecord, 0, 2)
		for _, precord := range clp.logTransformer.Preprocess(logRec, clp.logBuffer) {
			clp.logBuffer.AddRecord(precord)
			rec, err := clp.logTransformer.Transform(precord, clp.appType, tzShiftMin, clp.anonymousUsers)
			ans = append(ans, rec)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to transform item %s", precord)
				return []conversion.OutputRecord{}
			}
			applyLocation(precord, clp.geoIPDb, rec)
		}
		return ans
	}
	clp.numNonLoggable++
	return []conversion.OutputRecord{}
}

// GetAppType returns a string idenfier unique for a concrete application we
// want to archive logs for (e.g. 'kontext', 'syd', ...)
func (clp *CNKLogProcessor) GetAppType() string {
	return clp.appType
}

// GetAppVersion returns an application version (major and minor version info, e.g. 0.15, 1.7)
func (clp *CNKLogProcessor) GetAppVersion() string {
	return clp.appVersion
}

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
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}
	userMap := users.EmptyUserMap()
	confPath := filepath.Join(conf.CustomConfDir, "usermap.json")
	if fsop.IsFile(confPath) {
		userMap, err = users.LoadUserMap(confPath)
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
	defer geoDb.Close()

	clientTypeDetector, err := botwatch.NewClientTypeAnalyzer(conf.BotDetection)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	finishEvent := make(chan bool)
	go func() {
		switch action {
		case config.ActionBatch:
			runBatchAction(conf, options, geoDb, userMap, clientTypeDetector, finishEvent)

		case config.ActionTail:
			runTailAction(conf, geoDb, userMap, options.dryRun, clientTypeDetector, finishEvent)
		}
	}()
	<-finishEvent

}
