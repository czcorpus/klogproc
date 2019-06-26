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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/czcorpus/klogproc/fetch/sfiles"
	"github.com/czcorpus/klogproc/fetch/sredis"

	"os"

	"github.com/czcorpus/klogproc/elastic"
	"github.com/czcorpus/klogproc/influx"
)

// Conf describes klogproc's configuration
type Conf struct {
	LogRedis       sredis.RedisConf   `json:"logRedis"`
	LogFiles       sfiles.Conf        `json:"logFiles"`
	GeoIPDbPath    string             `json:"geoIpDbPath"`
	LocalTimezone  string             `json:"localTimezone"`
	AnonymousUsers int                `json:"anonymousUsers"`
	LogPath        string             `json:"logPath"`
	RecUpdate      elastic.DocUpdConf `json:"recordUpdate"`
	ElasticSearch  elastic.SearchConf `json:"elasticSearch"`
	InfluxDB       influx.Conf        `json:"influxDb"`
	AppType        string             `json:"appType"`
}

// UsesRedis tests whether the config contains Redis
// configuration. The function is happy once it finds
// a non empty address. Other values are not checked here
// (it is up to the client module to validate that).
func (c *Conf) UsesRedis() bool {
	return c.LogRedis.Address != ""
}

// HasInfluxOut tests whether an InfluxDB
// output is confgured
func (c *Conf) HasInfluxOut() bool {
	return c.InfluxDB.Server != ""
}

// HasElasticOut tests whether an ElasticSearch
// output is confgured
func (c *Conf) HasElasticOut() bool {
	return c.ElasticSearch.Server != ""
}

// TODO fix/update this
func validateConf(conf *Conf) {
	if conf.AppType == "" {
		log.Fatal("Application type not set")
	}
	if conf.HasElasticOut() {
		if conf.ElasticSearch.ScrollTTL == "" {
			log.Fatal("elasticScrollTtl must be a valid ElasticSearch scroll arg value (e.g. '2m', '30s')")
		}
		if conf.ElasticSearch.PushChunkSize == 0 {
			log.Fatal("elasticPushChunkSize is missing")
		}
	}
}

func updateRecords(conf *Conf) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordUpdate(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.Update, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Printf("Updated %d items\n", totalUpdated)

		} else {
			log.Fatal("Update error: ", err)
		}
	}
}

func removeKeyFromRecords(conf *Conf) {
	client := elastic.NewClient(&conf.ElasticSearch)
	for _, updConf := range conf.RecUpdate.Filters {
		totalUpdated, err := client.ManualBulkRecordKeyRemove(conf.ElasticSearch.Index, updConf,
			conf.RecUpdate.RemoveKey, conf.ElasticSearch.ScrollTTL, conf.RecUpdate.SearchChunkSize)
		if err == nil {
			log.Printf("Removed key %s from %d items\n", conf.RecUpdate.RemoveKey, totalUpdated)

		} else {
			log.Fatal("Update error: ", err)
		}
	}
}

func help(topic string) {
	if topic == "" {
		fmt.Print("Missing action to help with. Select one of the:\n\tcreate-index, extract-ngrams, search-service, search")
	}
	fmt.Printf("\n[%s]\n\n", topic)
	switch topic {
	case "proclogs":
		fmt.Println(helpTexts[0])
	case "docupdate":
		fmt.Println(helpTexts[1])
	default:
		fmt.Println("- no information available -")
	}
	fmt.Println()
}

func loadConfig(path string) *Conf {
	if path == "" {
		log.Fatal("Config path not specified")
	}
	rawData, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
	var conf Conf
	json.Unmarshal(rawData, &conf)
	if conf.LocalTimezone == "" {
		conf.LocalTimezone = "+02:00" // add Czech timezone by default
	}
	return &conf
}

func setup(confPath string) (*Conf, *os.File) {
	conf := loadConfig(confPath)
	validateConf(conf)

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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Klogproc - an utility for parsing and sending KonText/Bonito logs to ElasticSearch\n\nUsage:\n\t%s [options] [action] [config.json]\n\nAavailable actions:\n\tproclogs, docupdate, keyremove, help\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	var conf *Conf
	var logf *os.File

	switch flag.Arg(0) {
	case "help":
		help(flag.Arg(1))
	case "docupdate":
		conf, logf = setup(flag.Arg(1))
		updateRecords(conf)
	case "keyremove":
		conf, logf = setup(flag.Arg(1))
		removeKeyFromRecords(conf)
	case "proclogs":
		conf, logf = setup(flag.Arg(1))
		processLogs(conf)
	case "jsonize":
		conf, logf = setup(flag.Arg(1))
		jsonizeLogs(conf)
	default:
		fmt.Printf("Unknown action [%s]. Try -h for help\n", flag.Arg(0))
		os.Exit(1)
	}

	if logf != nil {
		logf.Close()
	}
}
