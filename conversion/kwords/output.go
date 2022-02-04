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

package kwords

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"klogproc/conversion"
)

func createID(rec *OutputRecord) string {
	rls := ""
	if rec.RefLength != nil {
		rls = strconv.Itoa(*rec.RefLength)
	}
	str := rec.Type + rec.Datetime + rec.IPAddress + rec.UserID + strconv.Itoa(rec.NumFiles) +
		rec.TargetInputType + strconv.Itoa(rec.TargetLength) + rec.Corpus + rls +
		strconv.FormatBool(rec.Pronouns) + strconv.FormatBool(rec.Prep) + strconv.FormatBool(rec.Con) +
		strconv.FormatBool(rec.Num) + strconv.FormatBool(rec.CaseInsensitive)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// OutputRecord represents polished, export ready record from Kwords log
type OutputRecord struct {
	ID              string `json:"-"`
	Type            string `json:"type"`
	time            time.Time
	Datetime        string                   `json:"datetime"`
	IPAddress       string                   `json:"ipAddress"`
	UserID          string                   `json:"userId"`
	IsAnonymous     bool                     `json:"isAnonymous"`
	IsQuery         bool                     `json:"isQuery"`
	NumFiles        int                      `json:"numFiles"`
	TargetInputType string                   `json:"targetInputType"`
	TargetLength    int                      `json:"targetLength"`
	Corpus          string                   `json:"string"`
	RefLength       *int                     `json:"refLength"`
	Pronouns        bool                     `json:"pronouns"`
	Prep            bool                     `json:"prep"`
	Con             bool                     `json:"con"`
	Num             bool                     `json:"num"`
	CaseInsensitive bool                     `json:"caseInsensitive"`
	GeoIP           conversion.GeoDataRecord `json:"geoip"`
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
