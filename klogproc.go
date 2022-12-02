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
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"klogproc/config"
	"klogproc/load/batch"
	"klogproc/save/elastic"
)

const (
	actionBatch     = "batch"
	actionTail      = "tail"
	actionRedis     = "redis"
	actionCelery    = "celery"
	actionKeyremove = "keyremove"
	actionDocupdate = "docupdate"
	actionHelp      = "help"
	actionVersion   = "version"

	startingServiceMsg = "INFO: ######################## Starting klogproc ########################"
)

var (
	version         string
	build           string
	gitCommit       string
	logLevelMapping = map[string]zerolog.Level{
		"debug":   zerolog.DebugLevel,
		"info":    zerolog.InfoLevel,
		"warning": zerolog.WarnLevel,
		"warn":    zerolog.WarnLevel,
		"error":   zerolog.ErrorLevel,
	}
)

func updateRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordUpdate(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.Update, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Info().Msgf("Updated %d items\n", totalUpdated)

		} else {
			log.Fatal().Err(err).Msg("Failed to update records")
		}
	}
}

func removeKeyFromRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordKeyRemove(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.RemoveKey, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Info().Msgf("Removed key %s from %d items\n", conf.RecUpdate.RemoveKey, totalUpdated)

		} else {
			log.Fatal().Err(err).Msgf("Failed to update records")
		}
	}
}

func help(topic string) {
	if topic == "" {
		fmt.Print("Missing action to help with. Select one of the:\n\tcreate-index, extract-ngrams, search-service, search")
	}
	fmt.Printf("\n[%s]\n\n", topic)
	switch topic {
	case actionBatch:
		fmt.Println(helpTexts[0])
	case actionTail:
		fmt.Println(helpTexts[1])
	case actionRedis:
		fmt.Println(helpTexts[2])
	case actionDocupdate:
		fmt.Println(helpTexts[3])
	default:
		fmt.Println("- no information available -")
	}
	fmt.Println()
}

func setupLog(path, level string) {
	lev, ok := logLevelMapping[level]
	if !ok {
		log.Fatal().Msgf("invalid logging level: %s", level)
	}
	zerolog.SetGlobalLevel(lev)
	if path != "" {
		logf, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal().Msgf("Failed to initialize log. File: %s", path)
		}
		log.Logger = log.Output(logf)

	} else {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		)
	}
}

func setup(confPath string) *config.Main {
	conf := config.Load(confPath)
	config.Validate(conf)
	llevel := "info"
	if conf.LogLevel != "" {
		llevel = conf.LogLevel
	}
	setupLog(conf.LogPath, llevel)
	return conf
}

func main() {
	procOpts := new(ProcessOptions)
	flag.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	flag.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")
	fromTimestamp := flag.String("from-time", "", "Batch process only the records with datetime greater or equal to this time (UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	toTimestamp := flag.String("to-time", "", "Batch process only the records with datetime less or equal to this UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	flag.BoolVar(&procOpts.analysisOnly, "analysis-only", false, "In batch mode, analyze logs for bots etc.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Klogproc - an utility for parsing and sending CNC app logs to ElasticSearch & InfluxDB\n\nUsage:\n\t%s [options] [action] [config.json]\n\nAavailable actions:\n\t%s\n\nOptions:\n",
			filepath.Base(os.Args[0]), strings.Join([]string{actionBatch, actionTail, actionRedis, actionDocupdate, actionKeyremove, actionHelp, actionVersion}, ", "))
		flag.PrintDefaults()
	}
	flag.Parse()

	var err error
	procOpts.datetimeRange, err = batch.NewDateTimeRange(fromTimestamp, toTimestamp)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	var conf *config.Main
	action := flag.Arg(0)

	switch action {
	case actionHelp:
		help(flag.Arg(1))
	case actionDocupdate:
		conf = setup(flag.Arg(1))
		updateRecords(conf, procOpts)
	case actionKeyremove:
		conf = setup(flag.Arg(1))
		removeKeyFromRecords(conf, procOpts)
	case actionBatch, actionTail, actionRedis:
		conf = setup(flag.Arg(1))
		log.Print(startingServiceMsg)
		processLogs(conf, action, procOpts)
	case actionCelery:
		conf = setup(flag.Arg(1))
		log.Print(startingServiceMsg)
		processCeleryStatus(conf)
	case actionVersion:
		fmt.Printf("Klogproc %s\nbuild date: %s\nlast commit: %s\n", version, build, gitCommit)
	default:
		fmt.Printf("Unknown action [%s]. Try -h for help\n", flag.Arg(0))
		os.Exit(1)
	}
}
