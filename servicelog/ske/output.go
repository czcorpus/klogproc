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

package ske

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

// isEntryQuery returns true if the action is one of
// those we consider "queries".
func isEntryQuery(action string) bool {
	ea := []string{"first", "wordlist", "wsketch", "thes", "wsdiff"}
	for _, item := range ea {
		if item == action {
			return true
		}
	}
	return false
}

// OutputRecord represents a polished version of SkE's access log.
type OutputRecord struct {
	ID          string `json:"-"`
	Type        string `json:"type"`
	Corpus      string `json:"corpus"`
	Subcorpus   string `json:"subcorpus"`
	Limited     bool   `json:"limited"`
	Action      string `json:"action"`
	Datetime    string `json:"datetime"`
	time        time.Time
	IPAddress   string                   `json:"ipAddress"`
	UserAgent   string                   `json:"userAgent"`
	UserID      string                   `json:"userId"`
	IsAnonymous bool                     `json:"isAnonymous"`
	IsQuery     bool                     `json:"isQuery"`
	GeoIP       servicelog.GeoDataRecord `json:"geoip,omitempty"`
	ProcTime    float64                  `json:"procTime"`
	// TODO
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

func (r *OutputRecord) SetTime(t time.Time) {
	r.Datetime = t.Format(time.RFC3339)
	r.time = t
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
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
	case "Limited":
		if tValue, ok := value.(lua.LBool); ok {
			if tValue == lua.LTrue {
				r.Limited = true
			}
			return nil
		}
	case "Action":
		if tValue, ok := value.(lua.LString); ok {
			r.Action = string(tValue)
			return nil
		}
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			r.Datetime = string(tValue)
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			r.IPAddress = string(tValue)
			return nil
		}
	case "UserAgent":
		if tValue, ok := value.(lua.LString); ok {
			r.UserAgent = string(tValue)
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
	case "IsQuery":
		r.IsQuery = value == lua.LTrue
		return nil
	case "ProcTime":
		if tValue, ok := value.(lua.LNumber); ok {
			r.ProcTime = float64(tValue)
			return nil
		}
	}
	return scripting.InvalidAttrError{Attr: name}
}

func (r *OutputRecord) GenerateDeterministicID() string {
	str := r.Type + r.Corpus + r.Subcorpus + strconv.FormatBool(r.Limited) + r.Action + r.Datetime + r.IPAddress + r.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// importCorpname imports corpname information out of URL-raw information.
func importCorpname(rawCorpname string) (string, bool) {
	items := strings.Split(rawCorpname, "/")
	if len(items) > 1 && items[0] == "omezeni" {
		return items[1], true
	}
	return rawCorpname, false
}
