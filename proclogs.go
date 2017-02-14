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
	"fmt"
	"log"

	"github.com/czcorpus/klogproc/elpush"
	"github.com/czcorpus/klogproc/logs"
	"github.com/oschwald/geoip2-golang"
)

type CNKLogProcessor struct {
	geoIPDb   *geoip2.Reader
	chunk     chan *elpush.CNKRecord
	chunkSize int
	currIdx   int
}

// ProcItem is a callback function called by log parser
func (clp *CNKLogProcessor) ProcItem(appType string, record *logs.LogRecord) {
	if record.AgentIsLoggable() {
		rec := elpush.New(record, appType)
		ip := record.GetClientIP()
		if ip != nil {
			city, err := clp.geoIPDb.City(ip)
			if err != nil {
				log.Printf("Failed to fetch GeoIP data for IP %s: %s", ip.String(), err)

			} else {
				rec.GeoIP.IP = ip.String()
				rec.GeoIP.CountryName = city.Country.Names["en"]
				rec.GeoIP.Latitude = float32(city.Location.Latitude)
				rec.GeoIP.Longitude = float32(city.Location.Longitude)
				rec.GeoIP.Location[0] = rec.GeoIP.Latitude
				rec.GeoIP.Location[1] = rec.GeoIP.Longitude
				rec.GeoIP.Timezone = city.Location.TimeZone
			}
		}
		clp.chunk <- rec
	}
}

func pushDataToElastic(data [][]byte) {
	fmt.Println("PUSH DATA TO ELASTIC:...")
	for i := 0; i < len(data); i += 2 {
		fmt.Println("M: ", string(data[i]))
		fmt.Println("D: ", string(data[i+1]))
	}
	fmt.Println("SENDING chunk...", len(data))
	// TODO
}

// ProcessLogs runs through all the logs found in configuration and matching
// some basic properties (it is a query, preferably from a human user etc.).
// The "producer" part of the processing runs in a separate goroutine while
// the main goroutine consumes values via a channel and after each
// n-th (conf.ElasticPushChunkSize) item it stores data to the ElasticSearch
// server.
func ProcessLogs(conf *Conf) {
	worklog, err := logs.LoadWorklog(conf.WorklogPath)
	if err != nil {
		panic(err)
	}
	last := worklog.FindLastRecord()
	fmt.Println(worklog, last)

	geoDb, err := geoip2.Open(conf.GeoIPDbPath)
	if err != nil {
		panic(err)
	}
	defer geoDb.Close()

	chunkChannel := make(chan *elpush.CNKRecord, conf.ElasticPushChunkSize*2)
	go func() {
		processor := &CNKLogProcessor{
			geoIPDb:   geoDb,
			chunk:     chunkChannel,
			chunkSize: conf.ElasticPushChunkSize,
		}

		files := logs.GetFilesInDir(conf.LogDir)
		for _, file := range files {
			p := logs.NewParser(file, conf.GeoIPDbPath, conf.LocalTimezone)
			p.Parse(last, conf.AppType, processor)
		}
		close(chunkChannel)
	}()

	i := 0
	data := make([][]byte, conf.ElasticPushChunkSize*2+i)
	for v := range chunkChannel {
		fmt.Println(v.ID)
		jsonData, err := v.ToJSON()
		jsonMeta := elpush.CNKRecordMeta{ID: v.ID, Type: v.Type, Index: conf.ElasticIndex}
		jsonMetaES, err2 := (&elpush.ElasticCNKRecordMeta{Index: jsonMeta}).ToJSON()
		if err == nil && err2 == nil {
			data[i] = jsonData
			data[i+1] = jsonMetaES
			i += 2

		} else {
			log.Print("Failed to encode item ", v.Datetime)
		}
		if i == conf.ElasticPushChunkSize*2 {
			pushDataToElastic(data)
			i = 0
		}
	}
	pushDataToElastic(data[:i])
}
