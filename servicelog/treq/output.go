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

package treq

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

// OutputRecord is an archive-ready Treq log record
type OutputRecord struct {
	ID          string `json:"-"`
	Type        string `json:"type"`
	time        time.Time
	IsAPI       bool                     `json:"isApi"`
	Datetime    string                   `json:"datetime"`
	QLang       string                   `json:"qLang"`
	SecondLang  string                   `json:"secondLang"`
	IPAddress   string                   `json:"ipAddress"`
	UserID      string                   `json:"userId"`
	IsAnonymous bool                     `json:"isAnonymous"`
	Corpus      string                   `json:"corpus"`
	Subcorpus   string                   `json:"subcorpus"`
	IsQuery     bool                     `json:"isQuery"`
	IsRegexp    bool                     `json:"isRegexp"`
	IsCaseInsen bool                     `json:"isCaseInsen"`
	IsMultiWord bool                     `json:"isMultiWord"`
	IsLemma     bool                     `json:"lemma"`
	GeoIP       servicelog.GeoDataRecord `json:"geoip,omitempty"`
}

// SetTime is defined for other treq variants
// so they can all share the same output rec. type
func (r *OutputRecord) SetTime(t time.Time) {
	r.Datetime = t.Format(time.RFC3339)
	r.time = t
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
	str := r.Type + strconv.FormatBool(r.IsAPI) + r.Corpus + r.Datetime + r.QLang + r.SecondLang + r.IPAddress +
		r.UserID + r.Subcorpus + strconv.FormatBool(r.IsQuery) + strconv.FormatBool(r.IsRegexp) +
		strconv.FormatBool(r.IsCaseInsen) + strconv.FormatBool(r.IsMultiWord) + strconv.FormatBool(r.IsLemma)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func (r *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	switch name {
	case "ID":
		if tValue, ok := value.(lua.LString); ok {
			r.ID = string(tValue)
			return nil
		}
	case "Type":
		if tValue, ok := value.(lua.LString); ok {
			r.Type = string(tValue)
			return nil
		}
	case "IsAPI":
		r.IsAPI = value == lua.LTrue
		return nil
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			r.SetTime(servicelog.ConvertDatetimeString(string(tValue)))
			return nil
		}
	case "QLang":
		if tValue, ok := value.(lua.LString); ok {
			r.QLang = string(tValue)
			return nil
		}
	case "SecondLang":
		if tValue, ok := value.(lua.LString); ok {
			r.SecondLang = string(tValue)
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			r.IPAddress = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			r.UserID = string(tValue)
			return nil
		}
	case "IsAnonymous":
		r.IsAnonymous = value == lua.LTrue
		return nil
	case "Corpus":
		if tValue, ok := value.(lua.LString); ok {
			r.Corpus = string(tValue)
			return nil
		}
	case "Subcorpus":
		if tValue, ok := value.(lua.LString); ok {
			r.Subcorpus = string(tValue)
			return nil
		}
	case "IsQuery":
		r.IsQuery = value == lua.LTrue
		return nil
	case "IsRegexp":
		r.IsRegexp = value == lua.LTrue
		return nil
	case "IsCaseInsen":
		r.IsCaseInsen = value == lua.LTrue
		return nil
	case "IsMultiWord":
		r.IsMultiWord = value == lua.LTrue
	case "IsLemma":
		r.IsLemma = value == lua.LTrue
	}
	return scripting.InvalidAttrError{Attr: name}
}
