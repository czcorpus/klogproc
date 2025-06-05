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
	"context"
	"klogproc/analysis"
	"klogproc/config"
	"klogproc/load/batch"
	"klogproc/logbuffer"
	"klogproc/notifications"
	"klogproc/save"
	"klogproc/save/elastic"
	"klogproc/servicelog"
	"klogproc/trfactory"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog/log"
)

// cnkLogProcessor imports parsed log records represented
// as InputRecord instances
type cnkLogProcessor struct {
	appType        string
	appVersion     string
	anonymousUsers []int
	geoIPDb        *geoip2.Reader
	chunkSize      int
	numNonLoggable int
	skipAnalysis   bool
	logTransformer servicelog.LogItemTransformer
	logBuffer      servicelog.ServiceLogBuffer
}

func (clp *cnkLogProcessor) recordIsLoggable(logRec servicelog.InputRecord) bool {
	return logRec.IsProcessable()
}

// ProcItem transforms input log record into an output format.
// In case an unsupported record is encountered, nil is returned.
func (clp *cnkLogProcessor) ProcItem(
	logRec servicelog.InputRecord,
) []servicelog.OutputRecord {
	if clp.recordIsLoggable(logRec) {
		ans := make([]servicelog.OutputRecord, 0, 2)
		prepInp, err := clp.logTransformer.Preprocess(logRec, clp.logBuffer)
		if err != nil {
			log.Error().
				Str("appType", clp.appType).
				Str("appVersion", clp.appVersion).
				Err(err).Msgf("Failed to transform item %s", logRec)
			return []servicelog.OutputRecord{}
		}
		for _, precord := range prepInp {
			clp.logBuffer.AddRecord(precord)
			rec, err := clp.logTransformer.Transform(precord)
			if err != nil {
				log.Error().
					Str("appType", clp.appType).
					Str("appVersion", clp.appVersion).
					Err(err).Msgf("Failed to transform item %s", logRec)
				return []servicelog.OutputRecord{}
			}
			applyLocation(precord, clp.geoIPDb, rec)
			ans = append(ans, rec)
		}
		return ans
	}
	clp.numNonLoggable++
	return []servicelog.OutputRecord{}
}

// GetAppType returns a string idenfier unique for a concrete application we
// want to archive logs for (e.g. 'kontext', 'syd', ...)
func (clp *cnkLogProcessor) GetAppType() string {
	return clp.appType
}

// GetAppVersion returns an application version (major and minor version info, e.g. 0.15, 1.7)
func (clp *cnkLogProcessor) GetAppVersion() string {
	return clp.appVersion
}

func runBatchAction(
	conf *config.Main,
	options *ProcessOptions,
	geoDB *geoip2.Reader,
	finishEvent chan<- bool,
) {
	// For debugging e-mail notification, you can pass `conf.EmailNotification`
	// as the first argument and use the "batch" mode to tune log processing.
	nullMailNot, _ := notifications.NewNotifier(nil, conf.ConomiNotification, conf.TimezoneLocation())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lt, err := trfactory.GetLogTransformer(
		conf.LogFiles,
		conf.AnonymousUsers,
		false,
		nullMailNot,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run batch action")
		return
	}

	var buffStorage servicelog.ServiceLogBuffer
	var stateFactory func() logbuffer.SerializableState
	if conf.LogFiles.Buffer != nil && conf.LogFiles.Buffer.BotDetection != nil {
		stateFactory = func() logbuffer.SerializableState {
			return &analysis.BotAnalysisState{
				PrevNums:          logbuffer.NewSampleWithReplac[int](20), // TODO hardcoded 20
				FullBufferIPProps: collections.NewConcurrentMap[string, analysis.SuspiciousReqCounter](),
			}
		}

	} else {
		stateFactory = func() logbuffer.SerializableState {
			return &analysis.SimpleAnalysisState{}
		}
	}

	if conf.LogFiles.Buffer != nil {
		buffStorage = logbuffer.NewStorage[servicelog.InputRecord, logbuffer.SerializableState](
			conf.LogFiles.Buffer,
			options.worklogReset,
			conf.LogFiles.LogBufferStateDir,
			conf.LogFiles.SrcPath,
			stateFactory,
		)

	} else {
		buffStorage = logbuffer.NewDummyStorage[servicelog.InputRecord, logbuffer.SerializableState](
			func() logbuffer.SerializableState {
				return &analysis.BotAnalysisState{
					PrevNums:          logbuffer.NewSampleWithReplac[int](20), // TODO hardcoded 20
					FullBufferIPProps: collections.NewConcurrentMap[string, analysis.SuspiciousReqCounter](),
				}
			},
		)
	}

	processor := &cnkLogProcessor{
		geoIPDb:        geoDB,
		chunkSize:      conf.ElasticSearch.PushChunkSize,
		appType:        conf.LogFiles.AppType,
		appVersion:     conf.LogFiles.Version,
		logTransformer: lt,
		anonymousUsers: conf.AnonymousUsers,
		skipAnalysis:   conf.LogFiles.SkipAnalysis,
		logBuffer:      buffStorage,
	}
	channelWriteES := make(chan *servicelog.BoundOutputRecord, conf.ElasticSearch.PushChunkSize*2)
	worklog := batch.NewWorklog(conf.LogFiles.WorklogPath)
	log.Info().Msgf("using worklog %s", conf.LogFiles.WorklogPath)
	if options.worklogReset {
		log.Printf("truncated worklog %v", worklog)
		err := worklog.Reset()
		if err != nil {
			log.Fatal().Msgf("unable to initialize worklog: %s", err)
		}
	}
	defer worklog.Save()

	wait := make(chan any)
	if options.dryRun || options.analysisOnly {
		wch := save.RunWriteConsumer(channelWriteES, !options.analysisOnly)
		go func() {
			for range wch {
			}
			wait <- struct{}{}
		}()
		log.Warn().Msg("using dry-run mode, output goes to stdout")

	} else {
		wch := elastic.RunWriteConsumer(conf.LogFiles.AppType, &conf.ElasticSearch, channelWriteES)
		go func() {
			for confirm := range wch {
				if confirm.Error != nil {
					log.Error().Err(confirm.Error).Msg("failed to save data to ElasticSearch database")
					// TODO
				}
			}
			wait <- struct{}{}
		}()
	}
	proc := batch.CreateLogFileProcFunc(ctx, processor, options.datetimeRange, channelWriteES)
	proc(conf.LogFiles, worklog.GetLastRecord())
	<-wait
	log.Info().Msgf("Ignored %d non-loggable entries (bots, static files etc.)", processor.numNonLoggable)
	stateData := buffStorage.GetStateData(time.Now())
	if stateData != nil && !reflect.ValueOf(stateData).IsNil() {
		log.Debug().Any("report", buffStorage.GetStateData(time.Now()).Report()).Msg("state report")
	}
	finishEvent <- true
}
