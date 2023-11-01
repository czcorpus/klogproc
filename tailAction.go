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
	"sync"

	"klogproc/config"
	"klogproc/email"
	"klogproc/load/alarm"
	"klogproc/load/batch"
	"klogproc/load/tail"
	"klogproc/logbuffer"
	"klogproc/logbuffer/analysis"
	"klogproc/save"
	"klogproc/save/elastic"
	"klogproc/save/influx"
	"klogproc/servicelog"
	"klogproc/trfactory"
	"klogproc/users"

	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog/log"
)

// -----

type tailProcessor struct {
	appType           string
	filePath          string
	version           string
	tzShift           int
	checkIntervalSecs int
	maxLinesPerCheck  int
	conf              *config.Main
	lineParser        batch.LineParser
	logTransformer    servicelog.LogItemTransformer
	geoDB             *geoip2.Reader
	anonymousUsers    []int
	elasticChunkSize  int
	influxChunkSize   int
	alarm             servicelog.AppErrorRegister
	analysis          chan<- servicelog.InputRecord
	logBuffer         servicelog.ServiceLogBuffer
	dryRun            bool
}

func (tp *tailProcessor) OnCheckStart() (tail.LineProcConfirmChan, *tail.LogDataWriter) {
	itemConfirm := make(tail.LineProcConfirmChan, 10)
	dataWriter := tail.LogDataWriter{
		Elastic: make(chan *servicelog.BoundOutputRecord, tp.elasticChunkSize*2),
		Influx:  make(chan *servicelog.BoundOutputRecord, tp.influxChunkSize),
		Ignored: make(chan save.IgnoredItemMsg),
	}

	go func() {
		var waitMergeEnd sync.WaitGroup
		waitMergeEnd.Add(3)
		if tp.dryRun {
			confirmChan1 := save.RunWriteConsumer(dataWriter.Elastic, false)
			go func() {
				for item := range confirmChan1 {
					itemConfirm <- item
				}
				waitMergeEnd.Done()
			}()
			confirmChan2 := save.RunWriteConsumer(dataWriter.Influx, false)
			go func() {
				for item := range confirmChan2 {
					itemConfirm <- item
				}
				waitMergeEnd.Done()
			}()
			log.Warn().Msg("using dry-run mode, output goes to stdout")

		} else {
			confirmChan1 := elastic.RunWriteConsumer(
				tp.appType, &tp.conf.ElasticSearch, dataWriter.Elastic)
			go func() {
				for item := range confirmChan1 {
					itemConfirm <- item
				}
				waitMergeEnd.Done()
			}()
			confirmChan2 := influx.RunWriteConsumer(
				&tp.conf.InfluxDB, dataWriter.Influx)
			go func() {
				for item := range confirmChan2 {
					itemConfirm <- item
				}
				waitMergeEnd.Done()
			}()
		}
		go func() {
			for msg := range dataWriter.Ignored {
				itemConfirm <- msg
			}
			waitMergeEnd.Done()
		}()
		waitMergeEnd.Wait()
		close(itemConfirm)
	}()

	return itemConfirm, &dataWriter
}

func (tp *tailProcessor) OnEntry(
	dataWriter *tail.LogDataWriter,
	item string,
	logPosition servicelog.LogRange,
) {
	parsed, err := tp.lineParser.ParseLine(item, -1) // TODO (line num - hard to keep track)
	if err != nil {
		switch tErr := err.(type) {
		case servicelog.LineParsingError:
			log.Warn().Err(tErr).Msgf("parsing error in file %s", tp.filePath)
		default:
			log.Error().Err(tErr).Send()
		}
		dataWriter.Ignored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
		return
	}
	if parsed.IsProcessable() {
		for _, precord := range tp.logTransformer.Preprocess(parsed, tp.logBuffer) {
			tp.logBuffer.AddRecord(precord)
			outRec, err := tp.logTransformer.Transform(precord, tp.appType, tp.tzShift, tp.anonymousUsers)
			if err != nil {
				log.Error().Err(err).Msg("Failed to transform processable record")
				dataWriter.Ignored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
				return
			}
			applyLocation(precord, tp.geoDB, outRec)
			dataWriter.Elastic <- &servicelog.BoundOutputRecord{
				FilePath: tp.filePath,
				Rec:      outRec,
				FilePos:  logPosition,
			}
			dataWriter.Influx <- &servicelog.BoundOutputRecord{
				FilePath: tp.filePath,
				Rec:      outRec,
				FilePos:  logPosition,
			}
		}

	} else {
		dataWriter.Ignored <- save.NewIgnoredItemMsg(tp.filePath, logPosition)
	}
}

func (tp *tailProcessor) OnCheckStop(dataWriter *tail.LogDataWriter) {
	close(dataWriter.Elastic)
	close(dataWriter.Influx)
	close(dataWriter.Ignored)
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

func (tp *tailProcessor) MaxLinesPerCheck() int {
	return tp.maxLinesPerCheck
}

// -----

func newProcAlarm(
	tailConf *tail.FileConf,
	conf *tail.Conf,
	notifier email.MailNotifier,
) (servicelog.AppErrorRegister, error) {
	if conf.NumErrorsAlarm > 0 && conf.ErrCountTimeRangeSecs > 0 && notifier != nil {
		return alarm.NewTailProcAlarm(
			conf.NumErrorsAlarm,
			conf.ErrCountTimeRangeSecs,
			tailConf,
			notifier,
		), nil
	}
	log.Warn().Msg("logged errors counting alarm not set")
	return &alarm.NullAlarm{}, nil
}

func newTailProcessor(
	tailConf tail.FileConf,
	conf config.Main,
	geoDB *geoip2.Reader,
	userMap *users.UserMap,
	logBuffers map[string]servicelog.ServiceLogBuffer,
	options *ProcessOptions,
) *tailProcessor {

	var notifier email.MailNotifier
	notifier, err := email.NewEmailNotifier(conf.EmailNotification, conf.TimezoneLocation())
	if err != nil {
		log.Fatal().Msgf("Failed to initialize e-mail notifier: %s", err)
	}

	procAlarm, err := newProcAlarm(&tailConf, conf.LogTail, notifier)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize alarm: %s", err)
	}
	lineParser, err := batch.NewLineParser(tailConf.AppType, tailConf.Version, procAlarm)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize parser: %s", err)
	}
	logTransformer, err := trfactory.GetLogTransformer(
		tailConf.AppType, tailConf.Version, tailConf.Buffer, userMap, true, notifier)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize transformer: %s", err)
	}
	log.Info().Msgf(
		"Creating tail processor for %s, app type: %s, app version: %s, tzShift: %d",
		filepath.Clean(tailConf.Path), tailConf.AppType, tailConf.Version, tailConf.TZShift)

	var buffStorage logbuffer.AbstractStorage[servicelog.InputRecord, logbuffer.SerializableState]
	if tailConf.Buffer != nil {
		var stateFactory func() logbuffer.SerializableState
		if tailConf.Buffer.BotDetection != nil {
			stateFactory = func() logbuffer.SerializableState {
				return &analysis.BotAnalysisState{
					PrevNums: logbuffer.NewSampleWithReplac[int](tailConf.Buffer.BotDetection.PrevNumReqsSampleSize),
				}
			}

		} else {
			stateFactory = func() logbuffer.SerializableState {
				return &analysis.SimpleAnalysisState{}
			}
		}

		if tailConf.Buffer.ID != "" {
			curr, ok := logBuffers[tailConf.Buffer.ID]
			if ok {
				log.Info().
					Str("bufferId", tailConf.Buffer.ID).
					Str("appType", tailConf.AppType).
					Str("file", tailConf.Path).
					Msg("reusing log processing buffer")
				buffStorage = curr

			} else {
				log.Info().
					Str("bufferId", tailConf.Buffer.ID).
					Str("appType", tailConf.AppType).
					Str("file", tailConf.Path).
					Msg("creating reusable log processing buffer")
				buffStorage = logbuffer.NewStorage[servicelog.InputRecord, logbuffer.SerializableState](
					tailConf.Buffer,
					options.worklogReset,
					conf.LogTail.LogBufferStateDir,
					tailConf.Path,
					stateFactory,
				)
				logBuffers[tailConf.Buffer.ID] = buffStorage
			}

		} else {
			buffStorage = logbuffer.NewStorage[servicelog.InputRecord, logbuffer.SerializableState](
				tailConf.Buffer,
				options.worklogReset,
				conf.LogTail.LogBufferStateDir,
				tailConf.Path,
				stateFactory,
			)
		}

	} else {
		buffStorage = logbuffer.NewDummyStorage[servicelog.InputRecord, logbuffer.SerializableState](
			func() logbuffer.SerializableState {
				return &analysis.BotAnalysisState{
					PrevNums: logbuffer.NewSampleWithReplac[int](tailConf.Buffer.BotDetection.PrevNumReqsSampleSize),
				}
			},
		)
	}

	return &tailProcessor{
		appType:           tailConf.AppType,
		filePath:          filepath.Clean(tailConf.Path), // note: this is not a full path normalization !
		version:           tailConf.Version,
		tzShift:           tailConf.TZShift,
		checkIntervalSecs: conf.LogTail.IntervalSecs,     // TODO maybe per-app type here ??
		maxLinesPerCheck:  conf.LogTail.MaxLinesPerCheck, // TODO dtto
		conf:              &conf,
		lineParser:        lineParser,
		logTransformer:    logTransformer,
		geoDB:             geoDB,
		anonymousUsers:    conf.AnonymousUsers,
		elasticChunkSize:  conf.ElasticSearch.PushChunkSize,
		influxChunkSize:   conf.InfluxDB.PushChunkSize,
		alarm:             procAlarm,
		logBuffer:         buffStorage,
		dryRun:            options.dryRun,
	}
}

// -----

func runTailAction(
	conf *config.Main,
	options *ProcessOptions,
	geoDB *geoip2.Reader,
	userMap *users.UserMap,
	finishEvt chan bool,
) {
	tailProcessors := make([]tail.FileTailProcessor, len(conf.LogTail.Files))
	var wg sync.WaitGroup
	wg.Add(len(conf.LogTail.Files))

	logBuffers := make(map[string]servicelog.ServiceLogBuffer)
	fullFiles, err := conf.LogTail.FullFiles()
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize files configuration")
		finishEvt <- true
		return
	}

	for i, f := range fullFiles {
		tailProcessors[i] = newTailProcessor(f, *conf, geoDB, userMap, logBuffers, options)
	}
	go func() {
		wg.Wait()
	}()
	go tail.Run(conf.LogTail, tailProcessors, finishEvt)
}
