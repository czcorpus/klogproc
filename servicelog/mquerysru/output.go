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

package mquerysru

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

// OutputRecord represents a polished version of WaG's access log.
type OutputRecord struct {
	ID        string `json:"-"`
	Type      string `json:"type"`
	Datetime  string `json:"datetime"`
	datetime  time.Time
	Level     string                   `json:"level"`
	IPAddress string                   `json:"ipAddress"`
	ProcTime  float64                  `json:"procTime"`
	Error     *servicelog.ErrorRecord  `json:"error,omitempty"`
	GeoIP     servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Corpus    string                   `json:"corpus,omitempty"`
	Version   string                   `json:"version"`
	Operation string                   `json:"operation"`
	IsQuery   bool                     `json:"isQuery"`
	Args      InputArgs                `json:"args"`
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

func (r *OutputRecord) SetTime(t time.Time) {
	r.Datetime = t.Format(time.RFC3339)
	r.datetime = t
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

func (r *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (r *OutputRecord) GenerateDeterministicID() string {
	str := r.Level + r.Datetime + r.IPAddress + r.Operation +
		strconv.FormatFloat(r.ProcTime, 'E', -1, 64) + r.Error.String()
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
