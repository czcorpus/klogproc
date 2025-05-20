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
	"fmt"
	"os"
	"time"

	"klogproc/common"
	"klogproc/fsop"
	"klogproc/load/batch"
	"klogproc/load/tail"
	"klogproc/save/elastic"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/cnc-gokit/mail"
	conomiClient "github.com/czcorpus/conomi/client"
	"github.com/rs/zerolog/log"
)

const (
	ActionBatch            = "batch"
	ActionTail             = "tail"
	ActionKeyremove        = "keyremove"
	ActionDocupdate        = "docupdate"
	ActionDocremove        = "docremove"
	ActionHelp             = "help"
	ActionMkScript         = "mkscript"
	ActionVersion          = "version"
	ActionTestNotification = "test-notification"

	DefaultTimeZone                       = "Europe/Prague"
	DefaultAlarmMaxLogInactivitySecs      = 3600 * 20
	DefaultLogInactivityCheckIntervalSecs = 3600
)

// Main describes klogproc's configuration
type Main struct {
	LogFiles                  *batch.Conf                    `json:"logFiles"`
	LogTail                   *tail.Conf                     `json:"logTail"`
	GeoIPDbPath               string                         `json:"geoIpDbPath"`
	AnonymousUsers            []int                          `json:"anonymousUsers"`
	Logging                   logging.LoggingConf            `json:"logging"`
	RecUpdate                 elastic.DocUpdConf             `json:"recordUpdate"`
	RecRemove                 elastic.DocRemConf             `json:"recordRemove"`
	ElasticSearch             elastic.ConnectionConf         `json:"elasticSearch"`
	EmailNotification         *mail.NotificationConf         `json:"emailNotification"`
	ConomiNotification        *conomiClient.ConomiClientConf `json:"conomiNotification"`
	TimeZone                  string                         `json:"timeZone"`
	AlarmMaxLogInactivitySecs int                            `json:"alarmMaxLogInactivitySecs"`

	// NotificationTag provides a better identification of a message source when sending
	// warnings to Conomi
	NotificationTag string `json:"notificationTag"`
}

func (c *Main) TimezoneLocation() *time.Location {
	// we can ignore the error here as we always call c.Validate()
	// first (which also tries to load the location and report possible
	// error)
	loc, _ := time.LoadLocation(c.TimeZone)
	return loc
}

// Validate checks for some essential config properties
// TODO test additional important items
func Validate(conf *Main, action string) {
	var err error
	if conf.ElasticSearch.IsConfigured() {
		err = conf.ElasticSearch.Validate()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to validate Elasticsearch configuration")
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
	}
	if conf.LogFiles != nil {
		if err := conf.LogFiles.Validate(); err != nil {
			log.Fatal().Err(err).Msg("logFiles validation error")
		}
	}
	if conf.TimeZone == "" {
		conf.TimeZone = DefaultTimeZone
		log.Warn().Str("timezone", conf.TimeZone).
			Msg("timeZone not specified, using default")
	}
	if conf.AlarmMaxLogInactivitySecs == 0 {
		conf.AlarmMaxLogInactivitySecs = DefaultAlarmMaxLogInactivitySecs
		log.Warn().
			Str("value", fmt.Sprintf("%v", time.Duration(conf.AlarmMaxLogInactivitySecs)*time.Second)).
			Msg("alarmMaxLogInactivitySecs not set, using default")
	}
	if conf.NotificationTag == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown-host"
		}
		conf.NotificationTag = fmt.Sprintf("unspecified instance at %s", hostname)
		log.Warn().
			Str("tag", conf.NotificationTag).
			Msg("notificationTag not specified, using default")
	}
}

// Load loads main configuration (either from a local fs or via http(s))
func Load(path string) *Main {
	rawData, err := common.LoadSupportedResource(path)
	if err != nil {
		log.Fatal().Err(err).Str("confSrc", path).Msgf("failed to load configuration")
	}
	var conf Main
	err = json.Unmarshal(rawData, &conf)
	if err != nil {
		log.Fatal().Err(err).Str("confSrc", path).Msgf("failed to unmarshal configuration")
	}
	return &conf
}
