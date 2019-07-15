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
	"time"
	"strconv"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	"github.com/czcorpus/klogproc/conversion"
)


func createID(rec *OutputRecord) string {
	str := rec.Type + rec.Datetime + rec.IPAddress + rec.UserID + rec.KeyReq + rec.KeyUsed +
		rec.Key + rec.RunScript + rec.Corpus + strconv.Itoa(rec.MinFreq) + rec.InputAttr + rec.OutputAttr +
		rec.CaseInsensitive
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

type OutputRecord struct {
	ID              string `json:"-"`
	Type            string `json:"type"`
	time            time.Time
	Datetime        string `json:"datetime"`
	IPAddress       string `json:"ipAddress"`
	UserID          string `json:"userId"`
	IsAnonymous	bool `json:"isAnonymous"`
	KeyReq          string `json:"keyReq"`
	KeyUsed         string `json:"keyUsed"`
	Key             string `json:"key"`
	RunScript       string `json:"runScript"`
	Corpus          string `json:"corpus`
	MinFreq         int    `json:"minFreq"`
	InputAttr       string `json:"inputAttr"`
	OutputAttr      string `json:"outputAttr"`
	CaseInsensitive string `json:"caseInsensitive"`
	GeoIP       conversion.GeoDataRecord `json:"geoip"`
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