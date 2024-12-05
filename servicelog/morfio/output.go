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

package morfio

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

// OutputRecord represents polished, export ready record from Morfio log
type OutputRecord struct {
	ID              string `json:"-"`
	Type            string `json:"type"`
	time            time.Time
	Datetime        string                   `json:"datetime"`
	IPAddress       string                   `json:"ipAddress"`
	UserID          string                   `json:"userId"`
	IsAnonymous     bool                     `json:"isAnonymous"`
	IsQuery         bool                     `json:"isQuery"`
	KeyReq          string                   `json:"keyReq"`
	KeyUsed         string                   `json:"keyUsed"`
	Key             string                   `json:"key"`
	RunScript       string                   `json:"runScript"`
	Corpus          string                   `json:"corpus"`
	MinFreq         int                      `json:"minFreq"`
	InputAttr       string                   `json:"inputAttr"`
	OutputAttr      string                   `json:"outputAttr"`
	CaseInsensitive bool                     `json:"caseInsensitive"`
	GeoIP           servicelog.GeoDataRecord `json:"geoip,omitempty"`
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

func (r *OutputRecord) GenerateDeterministicID() string {
	str := r.Type + r.Datetime + r.IPAddress + r.UserID + r.KeyReq + r.KeyUsed +
		r.Key + r.RunScript + r.Corpus + strconv.Itoa(r.MinFreq) + r.InputAttr + r.OutputAttr +
		strconv.FormatBool(r.CaseInsensitive)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func (r *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}
