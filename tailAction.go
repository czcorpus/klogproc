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
	"path/filepath"
	"sync"

	"klogproc/config"
	"klogproc/conversion"
	"klogproc/email"
	"klogproc/load/alarm"
	"klogproc/load/batch"
	"klogproc/load/tail"
	"klogproc/save"
	"klogproc/save/elastic"
	"klogproc/save/influx"
	"klogproc/users"

	"github.com/oschwald/geoip2-golang"
)

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
	dataForES         chan *conversion.BoundOutputRecord
	dataForInflux     chan *conversion.BoundOutputRecord
	dataIgnored       chan save.IgnoredItemMsg
	anonymousUsers    []int
	elasticChunkSize  int
	influxChunkSize   int
	itemConfirm       chan interface{}
	alarm             conversion.AppErrorRegister
	analysis          chan<- conversion.InputRecord
}

func (tp *tailProcessor) OnCheckStart() chan interface{} {
	tp.dataForES = make(chan *conversion.BoundOutputRecord, tp.elasticChunkSize*2)
	tp.dataForInflux = make(chan *conversion.BoundOutputRecord, tp.influxChunkSize)
	tp.dataIgnored = make(chan save.IgnoredItemMsg)

	var waitMergeEnd sync.WaitGroup
	waitMergeEnd.Add(3)
	tp.itemConfirm = make(chan interface{}, 10)
	confirmChan1 := elastic.RunWriteConsumer(tp.appType, &tp.conf.ElasticSearch, tp.dataForES)
	go func() {
		for item := range confirmChan1 {
			tp.itemConfirm <- item
		}
		waitMergeEnd.Done()
	}()
	confirmChan2 := influx.RunWriteConsumer(&tp.conf.InfluxDB, tp.dataForInflux)
	go func() {
		for item := range confirmChan2 {
			tp.itemConfirm <- item
		}
		waitMergeEnd.Done()
	}()
	go func() {
		for msg := range tp.dataIgnored {
			tp.itemConfirm <- msg
		}
		waitMergeEnd.Done()
	}()
	go func() {
		waitMergeEnd.Wait()
		close(tp.itemConfirm)
	}()
	return tp.itemConfirm
}

func (tp *tailProcessor) OnEntry(item string, logPosition conversion.LogRange) {
	parsed, err := tp.lineParser.ParseLine(item, -1) // TODO (line num - hard to keep track)
	if err != nil {
		switch tErr := err.(type) {
		case conversion.LineParsingError:
			log.Printf("INFO: file %s, %s", tp.filePath, tErr)
		default:
			log.Print("ERROR: ", tErr)
		}
		tp.dataIgnored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
		return
	}
	tp.analysis <- parsed
	if parsed.IsProcessable() {
		outRec, err := tp.logTransformer.Transform(parsed, tp.appType, tp.tzShift, tp.anonymousUsers)
		if err != nil {
			log.Printf("ERROR: %s", err)
			tp.dataIgnored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
			return
		}
		applyLocation(parsed, tp.geoDB, outRec)
		tp.dataForES <- &conversion.BoundOutputRecord{
			FilePath: tp.filePath,
			Rec:      outRec,
			FilePos:  logPosition,
		}
		tp.dataForInflux <- &conversion.BoundOutputRecord{
			FilePath: tp.filePath,
			Rec:      outRec,
			FilePos:  logPosition,
		}

	} else {
		tp.dataIgnored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
	}
}

func (tp *tailProcessor) OnCheckStop() {
	close(tp.dataForES)
	close(tp.dataForInflux)
	close(tp.dataIgnored)
	tp.alarm.Evaluate()
}

func (tp *tailProcessor) OnQuit() {
	tp.alarm.Reset()
	close(tp.analysis)
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

func newTailProcessor(
	tailConf tail.FileConf,
	conf config.Main,
	geoDB *geoip2.Reader,
	userMap *users.UserMap,
	analysis chan<- conversion.InputRecord,
) *tailProcessor {
	procAlarm, err := newProcAlarm(&tailConf, &conf.LogTail, &conf.EmailNotification)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize alarm: ", err)
	}
	lineParser, err := batch.NewLineParser(tailConf.AppType, tailConf.Version, procAlarm)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize parser: ", err)
	}
	logTransformer, err := GetLogTransformer(tailConf.AppType, tailConf.Version, userMap)
	if err != nil {
		log.Fatal("FATAL: Failed to initialize transformer: ", err)
	}
	log.Printf(
		"Creating tail processor for %s, app type: %s, app version: %s, tzShift: %d",
		filepath.Clean(tailConf.Path), tailConf.AppType, tailConf.Version, tailConf.TZShift)

	return &tailProcessor{
		appType:           tailConf.AppType,
		filePath:          filepath.Clean(tailConf.Path), // note: this is not a full path normalization !
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
		analysis:          analysis,
	}
}

// -----

func runTailAction(
	conf *config.Main,
	geoDB *geoip2.Reader,
	userMap *users.UserMap,
	analyzer ClientAnalyzer,
	finishEvt chan bool,
) {
	tailProcessors := make([]tail.FileTailProcessor, len(conf.LogTail.Files))
	var wg sync.WaitGroup
	wg.Add(len(conf.LogTail.Files))

	for i, f := range conf.LogTail.Files {
		tpAnalysis := make(chan conversion.InputRecord, 50)
		go func(items chan conversion.InputRecord) {
			for item := range items {
				analyzer.Add(item)
			}
			wg.Done()
		}(tpAnalysis)
		tailProcessors[i] = newTailProcessor(f, *conf, geoDB, userMap, tpAnalysis)
	}
	go func() {
		wg.Wait()
		analyzer.Close()
	}()
	go tail.Run(&conf.LogTail, tailProcessors, analyzer, finishEvt)
}
