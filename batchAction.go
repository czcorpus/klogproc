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
	"klogproc/analysis"
	"klogproc/config"
	"klogproc/load/batch"
	"klogproc/logbuffer"
	"klogproc/notifications"
	"klogproc/save"
	"klogproc/save/elastic"
	"klogproc/servicelog"
	"klogproc/trfactory"
	"klogproc/users"
	"reflect"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog/log"
)

func runBatchAction(
	conf *config.Main,
	options *ProcessOptions,
	geoDB *geoip2.Reader,
	userMap *users.UserMap,
	finishEvent chan<- bool,
) {
	// For debugging e-mail notification, you can pass `conf.EmailNotification`
	// as the first argument and use the "batch" mode to tune log processing.
	nullMailNot, _ := notifications.NewNotifier(nil, conf.ConomiNotification, conf.TimezoneLocation())
	lt, err := trfactory.GetLogTransformer(
		conf.LogFiles.AppType,
		conf.LogFiles.Version,
		conf.LogFiles.Buffer,
		userMap,
		conf.LogFiles.ExcludeIPList,
		false,
		nullMailNot,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run batch action")
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

	processor := &CNKLogProcessor{
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
	proc := batch.CreateLogFileProcFunc(processor, options.datetimeRange, channelWriteES)
	proc(conf.LogFiles, worklog.GetLastRecord())
	<-wait
	log.Info().Msgf("Ignored %d non-loggable entries (bots, static files etc.)", processor.numNonLoggable)
	stateData := buffStorage.GetStateData(time.Now())
	if stateData != nil && !reflect.ValueOf(stateData).IsNil() {
		log.Debug().Any("report", buffStorage.GetStateData(time.Now()).Report()).Msg("state report")
	}
	finishEvent <- true
}
