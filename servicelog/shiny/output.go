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

package shiny

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"klogproc/scripting"
	"klogproc/servicelog"
	"strconv"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// OutputRecord represents a log format as written
// to an archive database.
type OutputRecord struct {
	ID string `json:"-"`

	// Type here represents any application readable by this (= shiny) module
	// and defined within respective factories. To add a new shiny-based app
	// it should be enought to extend factories in parserFactory.go and
	// transformerFactory.go.
	Type        string `json:"type"`
	time        time.Time
	Datetime    string                   `json:"datetime"`
	IPAddress   string                   `json:"IPAddress"`
	User        string                   `json:"user"`
	UserID      string                   `json:"userId"`
	IsAnonymous bool                     `json:"isAnonymous"`
	Lang        string                   `json:"lang"`
	UserAgent   string                   `json:"userAgent"`
	GeoIP       servicelog.GeoDataRecord `json:"geoip,omitempty"`
	IsQuery     bool                     `json:"isQuery"`
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

func (r *OutputRecord) SetTime(t time.Time) {
	r.Datetime = t.Format(time.RFC3339)
	r.time = t
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (r *OutputRecord) GenerateDeterministicID() string {
	str := r.Type + strconv.Itoa(int(r.time.Unix())) + r.UserID + r.IPAddress + r.Lang + r.UserAgent
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
