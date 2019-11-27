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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
}

// Load loads main configuration (either from a local fs or via http(s))
func Load(path string) *Main {
	rawData, err := LoadSupportedResource(flag.Arg(1))
	if err != nil {
		log.Fatal("FATAL: ", err)
	}
	var conf Main
	json.Unmarshal(rawData, &conf)
	if conf.LocalTimezone == "" {
		conf.LocalTimezone = "+02:00" // add Czech timezone by default
	}
	return &conf
}

func loadHTTPResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Configuration resource loading error: %s (url: %s)", resp.Status, url)
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LoadSupportedResource loads raw byte data for Klogproc configuration.
// Allowed formats are:
// 1) http://..., https://...
// 2) file:/localhost/..., file:///...
// 3) /abs/fs/path, rel/fs/path
func LoadSupportedResource(uri string) ([]byte, error) {
	if uri == "" {
		return nil, fmt.Errorf("No resource (http, file) specified")
	}
	var rawData []byte
	var err error
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		rawData, err = loadHTTPResource(uri)

	} else if strings.HasPrefix(uri, "file:/localhost/") {
		rawData, err = ioutil.ReadFile(uri[len("file:/localhost/")-1:])

	} else if strings.HasPrefix(uri, "file:///") {
		rawData, err = ioutil.ReadFile(uri[len("file:///")-1:])

	} else { // we assume a common fs path
		rawData, err = ioutil.ReadFile(uri)
	}
	if err != nil {
		return nil, err
	}
	return rawData, nil
}
