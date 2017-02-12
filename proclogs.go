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
	"github.com/czcorpus/klogproc/elpush"
	"github.com/czcorpus/klogproc/logs"
	"github.com/oschwald/geoip2-golang"
)

type CNKLogProcessor struct {
	geoIPDb *geoip2.Reader
}

func (clp *CNKLogProcessor) ProcItem(appType string, record *logs.LogRecord) {
	rec := elpush.New(record, appType)
	ip := record.GetClientIP()
	city, err := clp.geoIPDb.City(ip)
	//fmt.Println("CITY: ", city.Country.Names["en"])
	if err != nil {
		// TODO create error record

	} else {
		rec.GeoIP.IP = ip.String()
		rec.GeoIP.CountryName = city.Country.Names["en"]
		rec.GeoIP.Latitude = float32(city.Location.Latitude)
		rec.GeoIP.Longitude = float32(city.Location.Longitude)
		rec.GeoIP.Location[0] = rec.GeoIP.Latitude
		rec.GeoIP.Location[1] = rec.GeoIP.Longitude
		rec.GeoIP.Timezone = city.Location.TimeZone
	}
	out, err := rec.ToJSON()
	if err == nil {
		fmt.Println("DATA ", string(out))

	} else {
		fmt.Println("ERROR: ", err)
	}
}

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
	processor := &CNKLogProcessor{geoIPDb: geoDb}
	files := logs.GetFilesInDir(conf.LogDir)
	for _, file := range files {
		p := logs.NewParser(file, conf.GeoIPDbPath, conf.LocalTimezone)
		p.Parse(last, conf.AppType, processor)
	}
}
