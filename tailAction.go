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
	"log"
	"sync"

	"github.com/czcorpus/klogproc/config"
	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/email"
	"github.com/czcorpus/klogproc/load/alarm"
	"github.com/czcorpus/klogproc/load/batch"
	"github.com/czcorpus/klogproc/load/tail"
	"github.com/czcorpus/klogproc/save/elastic"
	"github.com/czcorpus/klogproc/save/influx"
	"github.com/czcorpus/klogproc/users"
	"github.com/oschwald/geoip2-golang"
)

type notifyFailedChunks struct{}

func (n *notifyFailedChunks) RescueFailedChunks(chunk [][]byte) error {
	if len(chunk) > 0 {
		log.Print("ERROR: failed to insert a chunk of size ", len(chunk))
	}
	return nil
}

// -----

type tailProcessor struct {
	appType           string
	filePath          string
	version           string
	tzShift           int
	checkIntervalSecs int
	conf              *config.Main
	lineParser        batch.LineParser
	logTransformer    conversion.LogItemTransformer
	geoDB             *geoip2.Reader
	dataForES         chan conversion.OutputRecord
	dataForInflux     chan conversion.OutputRecord
	anonymousUsers    []int
	elasticChunkSize  int
	influxChunkSize   int
	outSync           sync.WaitGroup
	alarm             conversion.AppErrorRegister
}

func (tp *tailProcessor) OnCheckStart() {
	tp.dataForES = make(chan conversion.OutputRecord, tp.elasticChunkSize*2)
	tp.dataForInflux = make(chan conversion.OutputRecord, tp.influxChunkSize)
	tp.outSync = sync.WaitGroup{}
	tp.outSync.Add(2)
	go elastic.RunWriteConsumer(tp.appType, &tp.conf.ElasticSearch, tp.dataForES, &tp.outSync, &notifyFailedChunks{})
	go influx.RunWriteConsumer(&tp.conf.InfluxDB, tp.dataForInflux, &tp.outSync)
}

func (tp *tailProcessor) OnEntry(item string) {
	parsed, err := tp.lineParser.ParseLine(item, 0)
	if err != nil {
		switch tErr := err.(type) {
		case conversion.LineParsingError:
			log.Printf("INFO: file %s, %s", tp.filePath, tErr)
		default:
			log.Print("ERROR: ", tErr)
		}
		return
	}
	if parsed.IsProcessable() {
		outRec, err := tp.logTransformer.Transform(parsed, tp.appType, tp.tzShift, tp.anonymousUsers)
		if err != nil {
			log.Printf("ERROR: %s", err)
			return
		}
		applyLocation(parsed, tp.geoDB, outRec)
		tp.dataForES <- outRec
		tp.dataForInflux <- outRec
	}
}

func (tp *tailProcessor) OnCheckStop() {
	close(tp.dataForES)
	close(tp.dataForInflux)
	tp.outSync.Wait()
	tp.alarm.Evaluate()
}

func (tp *tailProcessor) OnQuit() {
	tp.alarm.Reset()
}

func (tp *tailProcessor) AppType() string {
	return tp.appType
}

func (tp *tailProcessor) FilePath() string {
	return tp.filePath
}

func (tp *tailProcessor) CheckIntervalSecs() int {
	return tp.checkIntervalSecs
}

// -----

func newProcAlarm(tailConf *tail.FileConf, conf *tail.Conf, mailConf *config.Email) (conversion.AppErrorRegister, error) {
	if conf.NumErrorsAlarm > 0 && conf.ErrCountTimeRangeSecs > 0 {
		notifier, err := email.NewEmailNotifier(mailConf)
		if err != nil {
			return nil, err
		}
		return alarm.NewTailProcAlarm(
			conf.NumErrorsAlarm,
			conf.ErrCountTimeRangeSecs,
			tailConf,
			notifier,
		), nil
	}
	log.Print("WARNING: logged errors counting alarm not set")
	return &alarm.NullAlarm{}, nil
}

func newTailProcessor(tailConf tail.FileConf, conf config.Main, geoDB *geoip2.Reader, userMap *users.UserMap) *tailProcessor {
	procAlarm, err := newProcAlarm(&tailConf, &conf.LogTail, &conf.EmailNotification)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize alarm: ", err)
	}
	lineParser, err := batch.NewLineParser(tailConf.AppType, procAlarm)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize parser: ", err)
	}
	logTransformer, err := GetLogTransformer(tailConf.AppType, tailConf.Version, userMap)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize transformer: ", err)
	}

	return &tailProcessor{
		appType:           tailConf.AppType,
		filePath:          tailConf.Path,
		version:           tailConf.Version,
		tzShift:           tailConf.TZShift,
		checkIntervalSecs: conf.LogTail.IntervalSecs, // TODO maybe per-app type here ??
		conf:              &conf,
		lineParser:        lineParser,
		logTransformer:    logTransformer,
		geoDB:             geoDB,
		anonymousUsers:    conf.AnonymousUsers,
		elasticChunkSize:  conf.ElasticSearch.PushChunkSize,
		influxChunkSize:   conf.InfluxDB.PushChunkSize,
		alarm:             procAlarm,
	}
}

// -----

func runTailAction(conf *config.Main, geoDB *geoip2.Reader, userMap *users.UserMap, clientAnalyzer ClientAnalyzer, finishEvt chan bool) {
	tailProcessors := make([]tail.FileTailProcessor, len(conf.LogTail.Files))
	for i, f := range conf.LogTail.Files {
		tailProcessors[i] = newTailProcessor(f, *conf, geoDB, userMap)

	}
	go tail.Run(&conf.LogTail, tailProcessors, clientAnalyzer, finishEvt)
}
