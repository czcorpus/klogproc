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
	"errors"
	"flag"

	"klogproc/botwatch"
	"klogproc/common"
	"klogproc/fsop"
	"klogproc/load/batch"
	"klogproc/load/celery"
	"klogproc/load/tail"
	"klogproc/save/elastic"
	"klogproc/save/influx"

	"github.com/rs/zerolog/log"
)

const (
	ActionBatch     = "batch"
	ActionTail      = "tail"
	ActionRedis     = "redis"
	ActionCelery    = "celery"
	ActionKeyremove = "keyremove"
	ActionDocupdate = "docupdate"
	ActionHelp      = "help"
	ActionVersion   = "version"
)

// Email configures e-mail client for sending misc. notifications and alarms
type Email struct {
	NotificationEmails []string `json:"notificationEmails"`
	SMTPServer         string   `json:"smtpServer"`
	Sender             string   `json:"sender"`
}

func (e *Email) Validate() error {
	if e == nil {
		return errors.New("missing whole e-mail notification section")
	}
	if len(e.NotificationEmails) == 0 {
		return errors.New("missing notification e-mails")
	}
	if e.SMTPServer == "" {
		return errors.New("missing SMTP server")
	}
	return nil
}

// Main describes klogproc's configuration
type Main struct {
	LogFiles          *batch.Conf               `json:"logFiles"`
	LogTail           *tail.Conf                `json:"logTail"`
	CeleryStatus      celery.Conf               `json:"celeryStatus"`
	GeoIPDbPath       string                    `json:"geoIpDbPath"`
	AnonymousUsers    []int                     `json:"anonymousUsers"`
	LogPath           string                    `json:"logPath"`
	LogLevel          string                    `json:"logLevel"`
	CustomConfDir     string                    `json:"customConfDir"`
	RecUpdate         elastic.DocUpdConf        `json:"recordUpdate"`
	ElasticSearch     elastic.ConnectionConf    `json:"elasticSearch"`
	InfluxDB          influx.ConnectionConf     `json:"influxDb"`
	EmailNotification *Email                    `json:"emailNotification"`
	BotDetection      botwatch.BotDetectionConf `json:"botDetection"`
}

// HasInfluxOut tests whether an InfluxDB
// output is confgured
func (c *Main) HasInfluxOut() bool {
	return c.InfluxDB.Server != ""
}

// Validate checks for some essential config properties
// TODO test additional important items
func Validate(conf *Main, action string) {
	var err error
	if conf.ElasticSearch.IsConfigured() {
		err = conf.ElasticSearch.Validate()
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
	if conf.InfluxDB.IsConfigured() {
		err = conf.InfluxDB.Validate()
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
	if !fsop.IsFile(conf.GeoIPDbPath) {
		log.Fatal().Msgf("Invalid GeoIPDbPath: '%s'", conf.GeoIPDbPath)
	}
	if action == ActionBatch && conf.LogFiles == nil {
		log.Fatal().Msg("missing configuration data for the `batch` action")
	}
	if action == ActionTail && conf.LogTail == nil {
		log.Fatal().Msg("missing configuration data for the `tail` action")
	}
	if conf.LogTail != nil {
		err := conf.LogTail.Validate()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to validate `tail` action configuration")
		}
		if conf.LogTail.RequiresMailConfiguration() {
			if err := conf.EmailNotification.Validate(); err != nil {
				log.Fatal().Err(err).Msg("failed to validate `tail` action configuration")
			}
		}
	}
	if conf.LogFiles != nil {
		if err := conf.LogFiles.Validate(); err != nil {
			log.Fatal().Err(err).Msg("logFiles validation error")
		}
	}
}

// Load loads main configuration (either from a local fs or via http(s))
func Load(path string) *Main {
	rawData, err := common.LoadSupportedResource(flag.Arg(1))
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}
	var conf Main
	json.Unmarshal(rawData, &conf)
	if conf.BotDetection.NumRequestsThreshold == 0 {
		log.Warn().Msg("botDetection.nmRequestsThreshold not set - using default 100")
		conf.BotDetection.NumRequestsThreshold = 100
	}
	return &conf
}
