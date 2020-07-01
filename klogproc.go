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
	"log"
	"path/filepath"
	"strings"

	"os"

	"github.com/czcorpus/klogproc/config"
	"github.com/czcorpus/klogproc/save/elastic"
)

const (
	actionBatch     = "batch"
	actionTail      = "tail"
	actionRedis     = "redis"
	actionCelery    = "celery"
	actionKeyremove = "keyremove"
	actionDocupdate = "docupdate"
	actionHelp      = "help"

	startingServiceMsg = "INFO: ######################## Starting klogproc ########################"
)

func updateRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordUpdate(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.Update, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Printf("INFO: Updated %d items\n", totalUpdated)

		} else {
			log.Fatal("FATAL: Update error: ", err)
		}
	}
}

func removeKeyFromRecords(conf *config.Main, options *ProcessOptions) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordKeyRemove(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.RemoveKey, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Printf("INFO: Removed key %s from %d items\n", conf.RecUpdate.RemoveKey, totalUpdated)

		} else {
			log.Fatal("FATAL: Update error: ", err)
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

func setup(confPath string) (*config.Main, *os.File) {
	conf := config.Load(confPath)
	config.Validate(conf)

	if conf.LogPath != "" {
		logf, err := os.OpenFile(conf.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to initialize log. File: %s", conf.LogPath)
		}
		log.SetOutput(logf)
		return conf, logf
	}
	return conf, nil
}

func main() {
	procOpts := new(ProcessOptions)
	flag.BoolVar(&procOpts.dryRun, "dry-run", false, "Do not write data (only for manual updates - batch, docupdate, keyremove)")
	flag.BoolVar(&procOpts.worklogReset, "worklog-reset", false, "Use the provided worklog but reset it first")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Klogproc - an utility for parsing and sending CNC app logs to ElasticSearch & InfluxDB\n\nUsage:\n\t%s [options] [action] [config.json]\n\nAavailable actions:\n\t%s\n\nOptions:\n",
			filepath.Base(os.Args[0]), strings.Join([]string{actionBatch, actionTail, actionRedis, actionDocupdate, actionKeyremove, actionHelp}, ", "))
		flag.PrintDefaults()
	}
	flag.Parse()

	var conf *config.Main
	var logf *os.File
	action := flag.Arg(0)

	switch action {
	case actionHelp:
		help(flag.Arg(1))
	case actionDocupdate:
		conf, logf = setup(flag.Arg(1))
		updateRecords(conf, procOpts)
	case actionKeyremove:
		conf, logf = setup(flag.Arg(1))
		removeKeyFromRecords(conf, procOpts)
	case actionBatch, actionTail, actionRedis:
		conf, logf = setup(flag.Arg(1))
		log.Print(startingServiceMsg)
		processLogs(conf, action, procOpts)
	case actionCelery:
		conf, logf = setup(flag.Arg(1))
		log.Print(startingServiceMsg)
		processCeleryStatus(conf)
	default:
		fmt.Printf("Unknown action [%s]. Try -h for help\n", flag.Arg(0))
		os.Exit(1)
	}

	if logf != nil {
		logf.Close()
	}
}
