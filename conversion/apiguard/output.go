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
	"encoding/json"
	"klogproc/conversion"
	"time"
)

type OutputRecord struct {
	InputRecord
	ID       string                   `json:"-"`
	GeoIP    conversion.GeoDataRecord `json:"geoip"`
	datetime time.Time
}

// ToJSON converts self to JSON string
func (orec *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(orec)
}

func (cnkr *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	tags = make(map[string]string)
	values = make(map[string]interface{})
	values["procTime"] = cnkr.ProcTime
	values["isCached"] = cnkr.IsCached
	values["isIndirect"] = cnkr.IsIndirect
	tags["type"] = cnkr.Type
	return
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

func (cnkr *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
	cnkr.GeoIP.IP = cnkr.IPAddress
	cnkr.GeoIP.CountryName = countryName
	cnkr.GeoIP.Latitude = latitude
	cnkr.GeoIP.Longitude = longitude
	cnkr.GeoIP.Location[0] = cnkr.GeoIP.Longitude
	cnkr.GeoIP.Location[1] = cnkr.GeoIP.Latitude
	cnkr.GeoIP.Timezone = timezone
}
