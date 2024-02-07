// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package kwords2

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"klogproc/servicelog"
	"time"
)

func createID(rec *OutputRecord) string {
	var uid string
	if rec.UserID != nil {
		uid = *rec.UserID
	}
	str := rec.Type + rec.Datetime + rec.Action + rec.IPAddress + uid + rec.Corpus
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

type Args struct {
	Attrs        []string `json:"attrs"`
	Level        float64  `json:"level"`
	EffectMetric string   `json:"effectMetric"`
	MinFreq      int      `json:"minFreq"`
	Percent      int      `json:"percent"`
}

// OutputRecord represents polished, export ready record from Kwords log
type OutputRecord struct {
	ID            string `json:"-"`
	Type          string `json:"type"`
	time          time.Time
	Datetime      string                   `json:"datetime"`
	IPAddress     string                   `json:"ipAddress"`
	UserID        *string                  `json:"userId"`
	IsAnonymous   bool                     `json:"isAnonymous"`
	Action        string                   `json:"action,omitempty"`
	IsQuery       bool                     `json:"isQuery"`
	Corpus        string                   `json:"corpus"`
	TextCharCount int                      `json:"textCharCount,omitempty"`
	TextWordCount int                      `json:"textWordCount,omitempty"`
	TextLang      string                   `json:"textLang,omitempty"`
	GeoIP         servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Args          *Args                    `json:"args,omitempty"`
	UserAgent     string                   `json:"userAgent"`
	Error         string                   `json:"error,omitempty"`
	Version       string                   `json:"version,omitempty"`
}

// SetLocation sets all the location related properties
func (r *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
	r.GeoIP.IP = r.IPAddress
	r.GeoIP.CountryName = countryName
	r.GeoIP.Latitude = latitude
	r.GeoIP.Longitude = longitude
	r.GeoIP.Location[0] = r.GeoIP.Longitude
	r.GeoIP.Location[1] = r.GeoIP.Latitude
	r.GeoIP.Timezone = timezone
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToInfluxDB creates tags and values to store in InfluxDB
func (r *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	return make(map[string]string), make(map[string]interface{})
}

// GetID Returns an unique ID of the record
func (r *OutputRecord) GetID() string {
	return r.ID
}

// GetType returns application type identifier
func (r *OutputRecord) GetType() string {
	return r.Type
}

// GetTime returns a creation time of the record
func (r *OutputRecord) GetTime() time.Time {
	return r.time
}
