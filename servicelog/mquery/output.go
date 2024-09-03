// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package mquery

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"klogproc/servicelog"
	"strconv"
	"time"
)

// OutputRecord represents a polished version of WaG's access log.
type OutputRecord struct {
	ID        string `json:"-"`
	Type      string `json:"type"`
	Datetime  string `json:"datetime"`
	datetime  time.Time
	Level     string                   `json:"level"`
	IPAddress string                   `json:"ipAddress"`
	UserAgent string                   `json:"userAgent"`
	IsAI      bool                     `json:"isAI"`
	ProcTime  float64                  `json:"procTime"`
	Error     *servicelog.ErrorRecord  `json:"error,omitempty"`
	GeoIP     servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Action    string                   `json:"action,omitempty"`
	CorpusID  string                   `json:"corpus,omitempty"`
}

// GetID returns an idempotent ID of the record.
func (r *OutputRecord) GetID() string {
	return r.ID
}

// GetType returns application type identifier
func (r *OutputRecord) GetType() string {
	return r.Type
}

// GetTime returns a creation time of the record
func (r *OutputRecord) GetTime() time.Time {
	return r.datetime
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
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

// CreateID creates an idempotent ID of rec based on its properties.
func CreateID(rec *OutputRecord) string {
	str := rec.Level + rec.Datetime + rec.IPAddress + rec.UserAgent + rec.CorpusID + rec.Action +
		strconv.FormatFloat(rec.ProcTime, 'E', -1, 64) + rec.Error.String()
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
