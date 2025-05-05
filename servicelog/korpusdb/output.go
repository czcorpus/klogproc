// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package korpusdb

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

// OutputRecord represents polished, export ready record from KorpusDB log
type OutputRecord struct {
	ID           string `json:"-"`
	Type         string `json:"type"`
	time         time.Time
	Path         string                   `json:"path"`
	Page         Pagination               `json:"page"`
	Datetime     string                   `json:"datetime"`
	IPAddress    string                   `json:"ipAddress"`
	UserID       string                   `json:"userId"`
	IsAnonymous  bool                     `json:"isAnonymous"`
	IsQuery      bool                     `json:"isQuery"`
	IsAPI        bool                     `json:"isApi"`
	IsPhraseBank bool                     `json:"isPhraseBank"`
	ClientFlag   string                   `json:"clientFlag"`
	GeoIP        servicelog.GeoDataRecord `json:"geoip,omitempty"`
	QueryType    string                   `json:"queryType"` // token/ngram
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

func (rec *OutputRecord) GenerateDeterministicID() string {
	str := rec.Type + rec.Path + rec.Datetime + rec.IPAddress + rec.UserID + rec.QueryType +
		strconv.Itoa(rec.Page.From) + strconv.Itoa(rec.Page.Size)
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
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			r.time = servicelog.ConvertDatetimeString(string(tValue))
			r.Datetime = string(tValue)
			return nil
		}
	case "Path":
		if tValue, ok := value.(lua.LString); ok {
			r.Path = string(tValue)
			return nil
		}
	case "Page":
		if tValue, ok := value.(*lua.LTable); ok {
			fromVal := tValue.RawGetString("From")
			if tFromVal, ok := fromVal.(lua.LNumber); ok {
				r.Page.From = int(tFromVal)
			}
			sizeVal := tValue.RawGetString("Size")
			if tSizeVal, ok := sizeVal.(lua.LNumber); ok {
				r.Page.Size = int(tSizeVal)
			}
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
	case "ClientFlag":
		if tValue, ok := value.(lua.LString); ok {
			r.ClientFlag = string(tValue)
			return nil
		}
	case "IsAnonymous":
		r.IsAnonymous = value == lua.LTrue
		return nil
	case "IsQuery":
		r.IsQuery = value == lua.LTrue
		return nil
	case "QueryType":
		if tValue, ok := value.(lua.LString); ok {
			r.QueryType = string(tValue)
			return nil
		}
	case "IsPhraseBank":
		r.IsAnonymous = value == lua.LTrue
		return nil
	}
	return scripting.InvalidAttrError{Attr: name}
}
