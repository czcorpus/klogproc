// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
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
	"github.com/rs/zerolog/log"

	"klogproc/load/batch"
	"klogproc/servicelog"

	"github.com/oschwald/geoip2-golang"
)

func applyLocation(rec servicelog.InputRecord, db *geoip2.Reader, outRec servicelog.OutputRecord) {
	ip := rec.GetClientIP()
	if len(ip) > 0 {
		city, err := db.City(ip)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch GeoIP data for IP %s.", ip.String())

		} else {
			outRec.SetLocation(city.Country.Names["en"], float32(city.Location.Latitude),
				float32(city.Location.Longitude), city.Location.TimeZone)
		}
	}
}

type ProcessOptions struct {
	worklogReset  bool
	dryRun        bool
	analysisOnly  bool
	datetimeRange batch.DatetimeRange
	scriptPath    string
	appType       string
}
