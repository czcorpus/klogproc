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

	"os"

	"github.com/czcorpus/klogproc/elastic"
)

// RedisConf is a structure containing information
// about Redis database containing logs to be
// processed.
type RedisConf struct {
	Address  string `json:"address"`
	Database int    `json:"database"`
	QueueKey string `json:"queueKey"`
}

// Conf describes klogproc's configuration
type Conf struct {
	WorklogPath                 string                      `json:"worklogPath"`
	AppType                     string                      `json:"appType"`
	LogDir                      string                      `json:"logDir"`
	LogRedis                    RedisConf                   `json:"logRedis"`
	GeoIPDbPath                 string                      `json:"geoIpDbPath"`
	LocalTimezone               string                      `json:"localTimezone"`
	AnonymousUsers              int                         `json:"anonymousUsers"`
	ImportPartiallyMatchingLogs bool                        `json:"importPartiallyMatchingLogs"`
	AppLogPath                  string                      `json:"appLogPath"`
	Updates                     []elastic.APIFlagUpdateConf `json:"updates"`
	elastic.ElasticSearchConf
}

// GetESConf returns ElasticSearch configuration part
// of the config.
func (c *Conf) GetESConf() *elastic.ElasticSearchConf {
	return &elastic.ElasticSearchConf{
		ElasticServer:          c.ElasticServer,
		ElasticIndex:           c.ElasticIndex,
		ElasticSearchChunkSize: c.ElasticSearchChunkSize,
		ElasticPushChunkSize:   c.ElasticPushChunkSize,
		ElasticScrollTTL:       c.ElasticScrollTTL,
	}
}

// UsesRedis tests whether the config contains Redis
// configuration. The function is happy once it finds
// a non empty address. Other values are not checked here
// (it is up to the client module to validate that).
func (c *Conf) UsesRedis() bool {
	return c.LogRedis.Address != ""
}

func validateConf(conf *Conf) {
	if conf.ElasticSearchChunkSize < 1 {
		panic("elasticSearchChunkSize must be >= 1")
	}
	if conf.AppType == "" {
		panic("Application type not set")
	}
	if conf.ElasticScrollTTL == "" {
		panic("elasticScrollTtl must be a valid ElasticSearch scroll arg value (e.g. '2m', '30s')")
	}
	if conf.ElasticPushChunkSize == 0 {
		panic("elasticPushChunkSize is missing")
	}
}

func updateIsAPIStatus(conf *Conf) {
	client := elastic.NewClient(conf.ElasticServer, conf.ElasticIndex, conf.ElasticSearchChunkSize)
	for _, updConf := range conf.Updates {
		totalUpdated, err := client.BulkUpdateSetAPIFlag(conf.ElasticIndex, updConf, conf.ElasticScrollTTL)
		if err == nil {
			fmt.Printf("Updated %d items", totalUpdated)

		} else {
			fmt.Println("Update error: ", err)
		}

	}
}

func loadConfig(path string) *Conf {
	rawData, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	var conf Conf
	json.Unmarshal(rawData, &conf)
	return &conf
}

func showHelp() {
	fmt.Println(`
Available operations: setapiflag, proclogs, help.
...TODO...`)
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 1 && flag.Arg(0) == "help" {
		showHelp()

	} else if len(flag.Args()) == 2 {
		conf := loadConfig(flag.Arg(1))
		validateConf(conf)

		if conf.AppLogPath != "" {
			logf, err := os.OpenFile(conf.AppLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(fmt.Sprintf("Failed to initialize log. File: %s", conf.AppLogPath))
			}
			defer logf.Close()
			log.SetOutput(logf)
		}

		switch flag.Arg(0) {
		case "setapiflag":
			updateIsAPIStatus(conf)
		case "proclogs":
			ProcessLogs(conf)
		}

	} else {
		panic("Invalid arguments. Expected format: klogproc OPERATION CONF")
	}

}
