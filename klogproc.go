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

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"klogproc/config"
	"klogproc/load/batch"
	"klogproc/notifications"
	"klogproc/save/elastic"
)

const (
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

func removeRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, remConf := range conf.RecRemove.Filters {
		totalRemoved, err := client.ManualBulkRecordRemove(conf.ElasticSearch.Index, remConf,
			conf.ElasticSearch.ScrollTTL, conf.RecRemove.SearchChunkSize, options.dryRun)
		if err == nil {
			if options.dryRun {
				log.Info().Msgf("%d items would be removed", totalRemoved)

			} else {
				log.Info().Msgf("Removed %d items", totalRemoved)
			}

		} else {
			log.Fatal().Err(err).Msg("Failed to remove records")
		}
	}
}

func removeKeyFromRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordKeyRemove(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.RemoveKey, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Info().Msgf("Removed key %s from %d items", conf.RecUpdate.RemoveKey, totalUpdated)

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
	case config.ActionBatch:
		fmt.Println(helpTexts[0])
	case config.ActionTail:
		fmt.Println(helpTexts[1])
	case config.ActionDocupdate:
		fmt.Println(helpTexts[3])
	default:
		fmt.Println("- no information available -")
	}
	fmt.Println()
}

func setup(confPath, action string) *config.Main {
	conf := config.Load(confPath)
	config.Validate(conf, action)
	if conf.Logging.Level == "" {
		conf.Logging.Level = "info"
	}
	logging.SetupLogging(conf.Logging)
	return conf
}

func main() {
	procOpts := new(ProcessOptions)

	batchCmd := flag.NewFlagSet(config.ActionBatch, flag.ExitOnError)
	batchCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	batchCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")
	fromTimestamp := batchCmd.String("from-time", "", "Batch process only the records with datetime greater or equal to this time (UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	toTimestamp := batchCmd.String("to-time", "", "Batch process only the records with datetime less or equal to this UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	batchCmd.BoolVar(&procOpts.analysisOnly, "analysis-only", false, "In batch mode, analyze logs for bots etc.")

	tailCmd := flag.NewFlagSet(config.ActionTail, flag.ExitOnError)
	tailCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")

	docupdateCmd := flag.NewFlagSet(config.ActionDocupdate, flag.ExitOnError)
	docupdateCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	docupdateCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")

	docremoveCmd := flag.NewFlagSet(config.ActionDocremove, flag.ExitOnError)
	docremoveCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	docremoveCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")

	keyremoveCmd := flag.NewFlagSet(config.ActionKeyremove, flag.ExitOnError)
	keyremoveCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	keyremoveCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")

	testnotifCmd := flag.NewFlagSet(config.ActionTestNotification, flag.ExitOnError)

	mkscriptCmd := flag.NewFlagSet(config.ActionMkScript, flag.ExitOnError)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Klogproc - an utility for parsing and sending CNC app logs to ElasticSearch\n\nUsage:\n\t%s [options] [action] [config.json]\n\nAavailable actions:\n\t%s\n\nOptions:\n",
			filepath.Base(
				os.Args[0]),
			strings.Join([]string{
				config.ActionBatch,
				config.ActionTail,
				config.ActionDocupdate,
				config.ActionKeyremove,
				config.ActionDocremove,
				config.ActionHelp,
				config.ActionVersion,
			}, ", "))
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
	case config.ActionHelp:
		help(flag.Arg(1))
	case config.ActionDocupdate:
		docupdateCmd.Parse(os.Args[2:])
		conf = setup(docupdateCmd.Arg(0), action)
		updateRecords(conf, procOpts)
	case config.ActionDocremove:
		docremoveCmd.Parse(os.Args[2:])
		conf = setup(docremoveCmd.Arg(0), action)
		removeRecords(conf, procOpts)
	case config.ActionKeyremove:
		keyremoveCmd.Parse(os.Args[2:])
		conf = setup(keyremoveCmd.Arg(0), action)
		removeKeyFromRecords(conf, procOpts)
	case config.ActionBatch:
		batchCmd.Parse(os.Args[2:])
		fmt.Println("PATH: ", batchCmd.Arg(0))
		fmt.Println("OPTS: ", procOpts)
		conf = setup(batchCmd.Arg(0), action)
		processLogs(conf, action, procOpts)
	case config.ActionTail:
		tailCmd.Parse(os.Args[2:])
		conf = setup(tailCmd.Arg(0), action)
		log.Print(startingServiceMsg)
		processLogs(conf, action, procOpts)
	case config.ActionTestNotification:
		testnotifCmd.Parse(os.Args[2:])
		conf = setup(testnotifCmd.Arg(0), action)
		notifier, err := notifications.NewNotifier(
			conf.EmailNotification, conf.ConomiNotification, conf.TimezoneLocation())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize notifier for testing")
		}
		notifier.SendNotification(
			"test",
			"Klogproc testing notification",
			map[string]any{"app": "klogproc", "dt": time.Now().In(conf.TimezoneLocation())},
			"This is just a testing notification triggered by running `klogproc test-notification`",
		)
	case config.ActionMkScript:
		mkscriptCmd.Parse(os.Args[2:])
		GenerateLuaStub(mkscriptCmd.Arg(0), mkscriptCmd.Arg(1))

	case config.ActionVersion:
		fmt.Printf("Klogproc %s\nbuild date: %s\nlast commit: %s\n", version, build, gitCommit)
	default:
		fmt.Printf("Unknown action [%s]. Try -h for help\n", flag.Arg(0))
		os.Exit(1)
	}
}
