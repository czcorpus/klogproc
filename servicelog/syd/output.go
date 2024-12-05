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

package syd

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"klogproc/scripting"
	"klogproc/servicelog"
	"strconv"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// OutputRecord represents a final format of log records for SyD as stored
// for further analysis and archiving
type OutputRecord struct {
	ID          string   `json:"-"`
	Type        string   `json:"type"`
	Corpus      []string `json:"corpus"`
	Datetime    string   `json:"datetime"`
	time        time.Time
	IPAddress   string `json:"ipAddress"`
	UserID      *int   `json:"userId"`
	IsAnonymous bool   `json:"isAnonymous"`
	KeyReq      string `json:"keyReq"`
	KeyUsed     string `json:"keyUsed"`
	Key         string `json:"key"`
	Ltool       string `json:"ltool"`
	RunScript   string `json:"runScript"`
	IsQuery     bool   `json:"isQuery"`

	GeoIP servicelog.GeoDataRecord `json:"geoip,omitempty"`
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

func (r *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (r *OutputRecord) GenerateDeterministicID() string {
	userID := "-"
	if r.UserID != nil {
		userID = strconv.Itoa(*r.UserID)
	}
	str := r.Type + strings.Join(r.Corpus, ":") + r.Datetime + r.IPAddress +
		userID + r.KeyReq + r.KeyUsed + r.Key + r.Ltool + r.RunScript +
		strconv.FormatBool(r.IsQuery)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
