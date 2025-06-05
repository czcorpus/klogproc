// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package kwords2

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"klogproc/scripting"
	"klogproc/servicelog"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type Args struct {
	Attrs        []string `json:"attrs"`
	Level        float64  `json:"level"`
	EffectMetric string   `json:"effectMetric"`
	MinFreq      int      `json:"minFreq"`
	Percent      int      `json:"percent"`
}

// OutputRecord represents polished, export ready record from Kwords log
type OutputRecord struct {
	ID            string `json:"-"`
	Type          string `json:"type"`
	time          time.Time
	Datetime      string                   `json:"datetime"`
	IPAddress     string                   `json:"ipAddress"`
	UserID        *string                  `json:"userId"`
	IsAnonymous   bool                     `json:"isAnonymous"`
	Action        string                   `json:"action,omitempty"`
	IsQuery       bool                     `json:"isQuery"`
	IsAPI         bool                     `json:"isApi"`
	Corpus        string                   `json:"corpus"`
	TextCharCount int                      `json:"textCharCount,omitempty"`
	TextWordCount int                      `json:"textWordCount,omitempty"`
	TextLang      string                   `json:"textLang,omitempty"`
	GeoIP         servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Args          *Args                    `json:"args,omitempty"`
	UserAgent     string                   `json:"userAgent"`
	Error         *servicelog.ErrorRecord  `json:"error,omitempty"`
	Version       string                   `json:"version,omitempty"`
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

func (r *OutputRecord) SetTime(t time.Time) {
	r.Datetime = t.Format(time.RFC3339)
	r.time = t
}

func (r *OutputRecord) GenerateDeterministicID() string {
	var uid string
	if r.UserID != nil {
		uid = *r.UserID
	}
	str := r.Type + r.Datetime + r.Action + r.IPAddress + uid + r.Corpus
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
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			r.IPAddress = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			v := string(tValue)
			r.UserID = &v
			return nil
		}
	case "IsAnonymous":
		if tValue, ok := value.(lua.LBool); ok {
			r.IsAnonymous = tValue == lua.LTrue
			return nil
		}
	case "Action":
		if tValue, ok := value.(lua.LString); ok {
			r.Action = string(tValue)
			return nil
		}
	case "IsQuery":
		if tValue, ok := value.(lua.LBool); ok {
			r.IsQuery = tValue == lua.LTrue
			return nil
		}
	case "IsAPI":
		if tValue, ok := value.(lua.LBool); ok {
			r.IsAPI = tValue == lua.LTrue
			return nil
		}
	case "Corpus":
		if tValue, ok := value.(lua.LString); ok {
			r.Corpus = string(tValue)
			return nil
		}
	case "TextCharCount":
		if tValue, ok := value.(lua.LNumber); ok {
			r.TextCharCount = int(tValue)
			return nil
		}
	case "TextWordCount":
		if tValue, ok := value.(lua.LNumber); ok {
			r.TextWordCount = int(tValue)
			return nil
		}
	case "TextLang":
		if tValue, ok := value.(lua.LString); ok {
			r.TextLang = string(tValue)
			return nil
		}
	case "Args":
		if tValue, ok := value.(*lua.LTable); ok {
			var err error
			tValue.ForEach(func(k, v lua.LValue) {
				tk, ok := k.(lua.LString)
				if !ok {
					err = fmt.Errorf("cannot set Args, key %v has a non-string type", k)
					return
				}
				switch string(tk) {
				case "Attrs": //         []string `json:"attrs"`
					tv, ok := v.(*lua.LTable)
					if !ok {
						err = fmt.Errorf("cannot set Args.Attrs - expected lua table")
						return
					}
					var tmp []string
					tmp, err = scripting.LuaTableToSliceOfStrings(tv)
					if err != nil {
						err = fmt.Errorf("cannot set Args.Attrs - failed to process values: %w", err)
						return
					}
					r.Args.Attrs = tmp
				case "Level":
					tv, ok := v.(lua.LNumber)
					if !ok {
						err = fmt.Errorf("canot set Args.Level - expected number")
						return
					}
					r.Args.Level = float64(tv)
				case "EffectMetric":
					tv, ok := v.(lua.LString)
					if !ok {
						err = fmt.Errorf("cannot set Args.EffectMetric - expected string")
					}
					r.Args.EffectMetric = string(tv)
				case "MinFreq":
					tv, ok := v.(lua.LNumber)
					if !ok {
						err = fmt.Errorf("canot set Args.MinFreq - expected number")
						return
					}
					r.Args.MinFreq = int(tv)
				case "Percent":
					tv, ok := v.(lua.LNumber)
					if !ok {
						err = fmt.Errorf("canot set Args.Percent - expected number")
						return
					}
					r.Args.Percent = int(tv)
				}
			})
			return nil
		}
	case "UserAgent":
		if tValue, ok := value.(lua.LString); ok {
			r.UserAgent = string(tValue)
			return nil
		}
	case "Error":
		if tValue, ok := value.(lua.LString); ok {
			r.Error = &servicelog.ErrorRecord{
				Name: string(tValue),
			}
			return nil
		}
	case "Version":
		if tValue, ok := value.(lua.LString); ok {
			r.Version = string(tValue)
			return nil
		}
	}
	return fmt.Errorf("invalid or read-only attribute %s", name)
}
