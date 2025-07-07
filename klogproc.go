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
	"bufio"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/oschwald/geoip2-golang"
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

func yesNoPrompt(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

func updateRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordUpdate(
			conf.ElasticSearch.Index,
			conf.RecUpdate.AppType,
			updConf,
			conf.RecUpdate.Update,
			conf.ElasticSearch.ScrollTTL,
			conf.RecUpdate.SearchChunkSize,
		)
		if err == nil {
			log.Info().Msgf("Updated %d items\n", totalUpdated)

		} else {
			log.Fatal().Err(err).Msg("Failed to update records")
		}
	}
}

func removeRecords(conf *config.Main, options *ProcessOptions) {
	if err := conf.RecRemove.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to remove records")
	}
	fmt.Fprintln(os.Stderr, "----------------------------------------------")
	fmt.Fprintf(os.Stderr, "the following subset(s) will be removed: \n")
	fmt.Fprintln(os.Stderr, conf.RecRemove.Overview())
	fmt.Fprintln(os.Stderr, "----------------------------------------------")
	var esclient *elastic.ESClient
	if conf.ElasticSearch.MajorVersion < 6 {
		esclient = elastic.NewClient(&conf.ElasticSearch)

	} else {
		esclient = elastic.NewClient6(&conf.ElasticSearch, conf.RecRemove.AppType)
	}
	count, err := esclient.CountRecords(conf.RecRemove.AppType, conf.RecRemove.Filters)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to count records")
	}
	if count == 0 {
		fmt.Fprintf(os.Stderr, "No matching records found.\n")
		return
	}
	if !yesNoPrompt(fmt.Sprintf("Found %d matching records. Are you sure to continue?", count), true) {
		return
	}
	log.Info().Msgf("%d items would be removed", count)
	for _, remConf := range conf.RecRemove.Filters {
		totalRemoved, err := esclient.ManualBulkRecordRemove(
			esclient.Index(),
			conf.RecRemove.AppType,
			remConf,
			conf.ElasticSearch.ScrollTTL,
			conf.RecRemove.SearchChunkSize,
			options.dryRun,
		)
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
	var esclient *elastic.ESClient
	if conf.ElasticSearch.MajorVersion < 6 {
		esclient = elastic.NewClient(&conf.ElasticSearch)

	} else {
		esclient = elastic.NewClient6(&conf.ElasticSearch, conf.RecRemove.AppType)
	}
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := esclient.ManualBulkRecordKeyRemove(
			esclient.Index(),
			conf.RecUpdate.AppType,
			updConf,
			conf.RecUpdate.RemoveKey,
			conf.ElasticSearch.ScrollTTL,
			conf.RecUpdate.SearchChunkSize,
		)
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
	if conf.Logging.Level == "" {
		conf.Logging.Level = "info"
	}
	logging.SetupLogging(conf.Logging)
	config.Validate(conf, action)
	return conf
}

func main() {
	procOpts := new(ProcessOptions)

	batchCmd := flag.NewFlagSet(config.ActionBatch, flag.ExitOnError)
	batchCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data anywhere, just print them")
	batchCmd.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")
	fromTimestamp := batchCmd.String("from-time", "", "Batch process only the records with datetime greater or equal to this time (UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	toTimestamp := batchCmd.String("to-time", "", "Batch process only the records with datetime less or equal to this UNIX timestamp, or YYYY-MM-DDTHH:mm:ss\u00B1hh:mm)")
	batchCmd.StringVar(&procOpts.scriptPath, "script-path", "", "Set or override Lua script path for log processing")
	noScript := batchCmd.Bool("no-script", false, "disables Lua script for log processing (overrides both cmd arg and json conf)")
	batchCmd.BoolVar(&procOpts.analysisOnly, "analysis-only", false, "In batch mode, analyze logs for bots etc.")

	tailCmd := flag.NewFlagSet(config.ActionTail, flag.ExitOnError)
	tailCmd.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data anywhere, just print them")
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
		fmt.Fprintf(os.Stderr, "Klogproc - an utility for processing CNC application logs\n\n"+
			"Usage:\n"+
			"\t%s batch [options] [config.json]\n"+
			"\t%s tail [options] [config.json]\n"+
			"\t%s docupdate [options] [config.json]\n"+
			"\t%s docremove [options] [config.json]\n"+
			"\t%s keyremove [options] [config.json]\n"+
			"\t%s test-nofification [options] [config.json]\n"+
			"\t%s mkscript [options] [config.json]\n"+
			"\t%s version\n",
			filepath.Base(os.Args[0]), filepath.Base(os.Args[0]), filepath.Base(os.Args[0]),
			filepath.Base(os.Args[0]), filepath.Base(os.Args[0]), filepath.Base(os.Args[0]),
			filepath.Base(os.Args[0]), filepath.Base(os.Args[0]))
	}
	flag.Parse()

	var err error
	procOpts.datetimeRange, err = batch.NewDateTimeRange(fromTimestamp, toTimestamp)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse command line date range")
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
		conf = setup(batchCmd.Arg(0), action)
		if *noScript {
			procOpts.scriptPath = ""
			conf.LogFiles.ScriptPath = ""

		} else if procOpts.scriptPath != "" {
			conf.LogFiles.ScriptPath = procOpts.scriptPath
		}
		geoDb, err := geoip2.Open(conf.GeoIPDbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open geo IP database")
		}
		defer geoDb.Close()
		finish := make(chan bool)
		go runBatchAction(conf, procOpts, geoDb, finish)
		<-finish
	case config.ActionTail:
		tailCmd.Parse(os.Args[2:])
		conf = setup(tailCmd.Arg(0), action)
		log.Print(startingServiceMsg)
		geoDb, err := geoip2.Open(conf.GeoIPDbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open geo IP database")
		}
		defer geoDb.Close()
		runTailAction(conf, procOpts, geoDb)
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
		if err := generateLuaStub(mkscriptCmd.Arg(0), mkscriptCmd.Arg(1)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case config.ActionVersion:
		fmt.Printf("Klogproc %s\nbuild date: %s\nlast commit: %s\n", version, build, gitCommit)
	default:
		fmt.Printf("Unknown action [%s]. Try -h for help\n", flag.Arg(0))
		os.Exit(1)
	}
}
