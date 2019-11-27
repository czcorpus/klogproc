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
	"io/ioutil"
	"log"

	"github.com/czcorpus/klogproc/load/batch"
	"github.com/czcorpus/klogproc/load/celery"
	"github.com/czcorpus/klogproc/load/sredis"
	"github.com/czcorpus/klogproc/load/tail"
	"github.com/czcorpus/klogproc/save/elastic"
	"github.com/czcorpus/klogproc/save/influx"
)

// Main describes klogproc's configuration
type Main struct {
	LogRedis       sredis.RedisConf       `json:"logRedis"`
	LogFiles       batch.Conf             `json:"logFiles"`
	LogTail        tail.Conf              `json:"logTail"`
	CeleryStatus   celery.Conf            `json:"celeryStatus"`
	GeoIPDbPath    string                 `json:"geoIpDbPath"`
	LocalTimezone  string                 `json:"localTimezone"`
	AnonymousUsers []int                  `json:"anonymousUsers"`
	LogPath        string                 `json:"logPath"`
	CustomConfDir  string                 `json:"customConfDir"`
	RecUpdate      elastic.DocUpdConf     `json:"recordUpdate"`
	ElasticSearch  elastic.ConnectionConf `json:"elasticSearch"`
	InfluxDB       influx.ConnectionConf  `json:"influxDb"`

	// BotDefsPath is either a local filesystem path or http resource path
	// where a list of bots to ignore etc. is defined
	BotDefsPath string `json:"botDefsPath"`
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
}

func Load(path string) *Main {
	if path == "" {
		log.Fatal("Config path not specified")
	}
	rawData, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
	var conf Main
	json.Unmarshal(rawData, &conf)
	if conf.LocalTimezone == "" {
		conf.LocalTimezone = "+02:00" // add Czech timezone by default
	}
	return &conf
}
