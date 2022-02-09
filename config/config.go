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

package config

import (
	"encoding/json"
	"flag"
	"log"

	"klogproc/botwatch"
	"klogproc/common"
	"klogproc/fsop"
	"klogproc/load/batch"
	"klogproc/load/celery"
	"klogproc/load/sredis"
	"klogproc/load/tail"
	"klogproc/save/elastic"
	"klogproc/save/influx"
)

// Email configures e-mail client for sending misc. notifications and alarms
type Email struct {
	NotificationEmails []string `json:"notificationEmails"`
	SMTPServer         string   `json:"smtpServer"`
	Sender             string   `json:"sender"`
}

// Main describes klogproc's configuration
type Main struct {
	LogRedis          sredis.RedisConf          `json:"logRedis"`
	LogFiles          batch.Conf                `json:"logFiles"`
	LogTail           tail.Conf                 `json:"logTail"`
	CeleryStatus      celery.Conf               `json:"celeryStatus"`
	GeoIPDbPath       string                    `json:"geoIpDbPath"`
	AnonymousUsers    []int                     `json:"anonymousUsers"`
	LogPath           string                    `json:"logPath"`
	CustomConfDir     string                    `json:"customConfDir"`
	RecUpdate         elastic.DocUpdConf        `json:"recordUpdate"`
	ElasticSearch     elastic.ConnectionConf    `json:"elasticSearch"`
	InfluxDB          influx.ConnectionConf     `json:"influxDb"`
	EmailNotification Email                     `json:"emailNotification"`
	BotDetection      botwatch.BotDetectionConf `json:"botDetection"`
}

// UsesRedis tests whether the config contains Redis
// configuration. The function is happy once it finds
// a non empty address. Other values are not checked here
// (it is up to the client module to validate that).
func (c *Main) UsesRedis() bool {
	return c.LogRedis.Address != ""
}

// HasInfluxOut tests whether an InfluxDB
// output is confgured
func (c *Main) HasInfluxOut() bool {
	return c.InfluxDB.Server != ""
}

// Validate checks for some essential config properties
// TODO test additional important items
func Validate(conf *Main) {
	var err error
	if conf.ElasticSearch.IsConfigured() {
		err = conf.ElasticSearch.Validate()
		if err != nil {
			log.Fatal("FATAL: ", err)
		}
	}
	if conf.InfluxDB.IsConfigured() {
		err = conf.InfluxDB.Validate()
		if err != nil {
			log.Fatal("FATAL: ", err)
		}
	}
	if !fsop.IsFile(conf.GeoIPDbPath) {
		log.Fatal("FATAL: Invalid GeoIPDbPath: '", conf.GeoIPDbPath, "'")
	}
}

// Load loads main configuration (either from a local fs or via http(s))
func Load(path string) *Main {
	rawData, err := common.LoadSupportedResource(flag.Arg(1))
	if err != nil {
		log.Fatal("FATAL: ", err)
	}
	var conf Main
	json.Unmarshal(rawData, &conf)
	return &conf
}
