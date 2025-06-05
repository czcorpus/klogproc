// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package apiguard

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"klogproc/scripting"
	"klogproc/servicelog"
	"strconv"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type OutputRecord struct {
	Type       string                   `json:"type"`
	IsQuery    bool                     `json:"isQuery"`
	Service    string                   `json:"service"`
	ProcTime   float64                  `json:"procTime"`
	IsCached   bool                     `json:"isCached"`
	IsIndirect bool                     `json:"isIndirect"`
	UserID     string                   `json:"userId"`
	IPAddress  string                   `json:"ipAddress,omitempty"`
	UserAgent  string                   `json:"userAgent,omitempty"`
	ID         string                   `json:"-"`
	GeoIP      servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Datetime   string                   `json:"datetime"`
	datetime   time.Time
}

// ToJSON converts self to JSON string
func (orec *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(orec)
}

func (cnkr *OutputRecord) GetID() string {
	return cnkr.ID
}

func (cnkr *OutputRecord) GetType() string {
	return cnkr.Type
}

// GetTime returns Go Time instance representing
// date and time when the record was created.
func (cnkr *OutputRecord) GetTime() time.Time {
	return cnkr.datetime
}

func (cnkr *OutputRecord) SetTime(t time.Time) {
	cnkr.Datetime = t.Format(time.RFC3339)
	cnkr.datetime = t
}

func (cnkr *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
	cnkr.GeoIP.IP = cnkr.IPAddress
	cnkr.GeoIP.CountryName = countryName
	cnkr.GeoIP.Latitude = latitude
	cnkr.GeoIP.Longitude = longitude
	cnkr.GeoIP.Location[0] = cnkr.GeoIP.Longitude
	cnkr.GeoIP.Location[1] = cnkr.GeoIP.Latitude
	cnkr.GeoIP.Timezone = timezone
}

func (cnkr *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (apgr *OutputRecord) GenerateDeterministicID() string {
	str := apgr.Datetime + strconv.FormatBool(apgr.IsQuery) + apgr.Service + apgr.Type +
		apgr.IPAddress + apgr.UserAgent + fmt.Sprintf("%01.3f", apgr.ProcTime) +
		strconv.FormatBool(apgr.IsCached) + strconv.FormatBool(apgr.IsIndirect)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
