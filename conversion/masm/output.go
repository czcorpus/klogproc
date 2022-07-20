// Copyright 2022 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2022 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2022 Institute of the Czech National Corpus,
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

package masm

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// OutputRecord represents a polished version of WaG's access log.
type OutputRecord struct {
	ID             string `json:"-"`
	Type           string `json:"type"`
	Level          string `json:"level"`
	Time           string `json:"time"`
	time           time.Time
	Message        string   `json:"message"`
	IsQuery        bool     `json:"isQuery,omitempty"`
	Corpus         string   `json:"corpus,omitempty"`
	AlignedCorpora []string `json:"alignedCorpora,omitempty"`
	IsAutocomplete bool     `json:"isAutocomplete,omitempty"`
	IsCached       bool     `json:"isCached,omitempty"`
	ProcTimeSecs   float64  `json:"procTimeSecs,omitempty"`
	Error          string   `json:"error,omitempty"`
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

// SetLocation sets all the location related properties
func (r *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {

}

// CreateID creates an idempotent ID of rec based on its properties.
func CreateID(rec *OutputRecord) string {
	str := rec.Level + rec.Time + rec.Message + strconv.FormatBool(rec.IsQuery) + rec.Corpus +
		strings.Join(rec.AlignedCorpora, ", ") + strconv.FormatBool(rec.IsAutocomplete) + strconv.FormatBool(rec.IsCached) +
		strconv.FormatFloat(rec.ProcTimeSecs, 'E', -1, 64) + rec.Error
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}
