// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package wag06

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"klogproc/conversion"
)

// OutputRecord represents a polished version of WaG's access log.
type OutputRecord struct {
	ID                  string `json:"-"`
	Type                string `json:"type"`
	Action              string `json:"action"`
	Datetime            string `json:"datetime"`
	time                time.Time
	IPAddress           string                   `json:"ipAddress"`
	UserAgent           string                   `json:"userAgent"`
	ReferringDomain     string                   `json:"referringDomain"`
	UserID              string                   `json:"userId"`
	IsAnonymous         bool                     `json:"isAnonymous"`
	IsQuery             bool                     `json:"isQuery"`
	IsMobileClient      bool                     `json:"isMobileClient"`
	HasPosSpecification bool                     `json:"hasPosSpecification"`
	QueryType           string                   `json:"queryType"`
	Lang1               string                   `json:"lang1"`
	Lang2               string                   `json:"lang2"`
	Queries             []string                 `json:"queries"`
	GeoIP               conversion.GeoDataRecord `json:"geoip,omitempty"`
	ProcTime            float32                  `json:"procTime"`
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
	return r.time
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToInfluxDB creates tags and values to store in InfluxDB
func (r *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	return make(map[string]string), make(map[string]interface{})
}

// NewTimedOutputRecord creates a general empty record with specified date and time
func NewTimedOutputRecord(t time.Time, tzShiftMin int) *OutputRecord {
	dt := t.Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339)
	if dt[len(dt)-1] == 'Z' {
		dt = dt[:len(dt)-1] + "+00:00"
	}
	return &OutputRecord{
		time:     t,
		Datetime: dt,
	}
}

// CreateID creates an idempotent ID of rec based on its properties.
func CreateID(rec *OutputRecord) string {
	str := rec.Type + rec.Action + strconv.FormatBool(rec.IsAnonymous) + rec.Datetime + rec.IPAddress +
		rec.UserID + rec.QueryType + strings.Join(rec.Queries, "--") + rec.Lang1 + rec.Lang2
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
