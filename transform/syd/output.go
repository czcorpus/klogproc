// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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
	"strconv"
	"strings"
	"time"

	"github.com/czcorpus/klogproc/transform"
)

func createID(rec *OutputRecord) string {
	str := rec.Datetime + strings.Join(rec.Corpus, ":") + rec.IPAddress +
		rec.UserID + rec.KeyReq + rec.KeyUsed + rec.Key + rec.Ltool + rec.RunScript +
		strconv.FormatBool(rec.IsQuery) + rec.Type + rec.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

type OutputRecord struct {
	ID        string
	Datetime  string
	time      time.Time
	IPAddress string
	UserID    string
	KeyReq    string
	KeyUsed   string
	Key       string
	Ltool     string
	RunScript string
	IsQuery   bool
	Corpus    []string
	Type      string
	GeoIP     transform.GeoDataRecord `json:"geoip"`
}

func (r *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
	r.GeoIP.IP = r.IPAddress
	r.GeoIP.CountryName = countryName
	r.GeoIP.Latitude = latitude
	r.GeoIP.Longitude = longitude
	r.GeoIP.Location[0] = r.GeoIP.Longitude
	r.GeoIP.Location[1] = r.GeoIP.Latitude
	r.GeoIP.Timezone = timezone
}

func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	return make(map[string]string), make(map[string]interface{})
}

func (r *OutputRecord) GetID() string {
	return r.ID
}

func (r *OutputRecord) GetType() string {
	return r.Type
}

func (r *OutputRecord) GetTime() time.Time {
	return r.time
}
