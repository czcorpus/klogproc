// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package kontext015

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"klogproc/scripting"
	"klogproc/servicelog"
	"net/url"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ImportQueryType translates KonText/Bonito query type argument
// into a more understandable form
func ImportQueryType(record KonTextInputRecord) string {
	return record.GetStringArg("qtype")
}

// ImportCorpname extracts actual corpus name from
// URL argument which may contain additional data (e.g. variant prefix)
func ImportCorpname(record KonTextInputRecord) string {
	var corpname string

	if record.HasArg("corpname") && record.GetStringArg("corpname") != "" {
		corpname = record.GetStringArg("corpname")
		corpname, _ = url.QueryUnescape(corpname)
		return corpname

	} else if record.HasArg("corpora") {
		c, ok := getSliceOfStrings(record.AllArgs(), "corpora")
		if ok && len(c) > 0 {
			return c[0]
		}
	}
	return ""
}

func IsEntryQuery(action string) bool {
	ea := []string{"/query_submit", "/wordlist/submit", "/pquery/freq_intersection"}
	for _, item := range ea {
		if item == action {
			return true
		}
	}
	return false
}

// ------------------- output rec. implementation ------------

// OutputRecord represents an exported application log record ready
// to be inserted into ElasticSearch index.
type OutputRecord struct {
	ID             string   `json:"-"`
	Type           string   `json:"type"`
	Action         string   `json:"action"`
	Corpus         string   `json:"corpus"`
	AlignedCorpora []string `json:"alignedCorpora"`
	Datetime       string   `json:"datetime"`
	time           time.Time
	IPAddress      string                   `json:"ipAddress"`
	IsAnonymous    bool                     `json:"isAnonymous"`
	IsAPI          bool                     `json:"isApi"`
	IsQuery        bool                     `json:"isQuery"`
	ProcTime       float64                  `json:"procTime"`
	QueryType      string                   `json:"queryType"`
	UserAgent      string                   `json:"userAgent"`
	UserID         string                   `json:"userId"`
	GeoIP          servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Error          *servicelog.ErrorRecord  `json:"error,omitempty"`
	Args           map[string]interface{}   `json:"args"`
}

// SetTime is defined for other treq variants
// so they can all share the same output rec. type
func (cnkr *OutputRecord) SetTime(t time.Time) {
	cnkr.Datetime = t.Format(time.RFC3339)
	cnkr.time = t
}

// ToJSON converts self to JSON string
func (cnkr *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(cnkr)
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
	return cnkr.time
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

func (cnkr *OutputRecord) GenerateDeterministicID() string {
	str := cnkr.Action + cnkr.Corpus + cnkr.Datetime + cnkr.IPAddress +
		cnkr.Type + cnkr.UserAgent + cnkr.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func (cnkr *OutputRecord) LSetProperty(name string, value lua.LValue) error {
	switch name {
	case "ID":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.ID = string(tValue)
			return nil
		}
	case "Type":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.Type = string(tValue)
			return nil
		}
	case "Action":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.Action = string(tValue)
			return nil
		}
	case "Corpus":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.Corpus = string(tValue)
			return nil
		}
	case "AlignedCorpora":
		if tValue, ok := value.(*lua.LTable); ok {
			var err error
			cnkr.AlignedCorpora, err = scripting.LuaTableToSliceOfStrings(tValue)
			return err
		}
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.time = servicelog.ConvertDatetimeString(string(tValue))
			cnkr.Datetime = string(tValue)
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.IPAddress = string(tValue)
			return nil
		}
	case "IsAnonymous":
		cnkr.IsAnonymous = value == lua.LTrue
		return nil
	case "IsAPI":
		cnkr.IsAPI = value == lua.LTrue
		return nil
	case "IsQuery":
		cnkr.IsQuery = value == lua.LTrue
		return nil
	case "ProcTime":
		if tValue, ok := value.(lua.LNumber); ok {
			cnkr.ProcTime = float64(tValue)
			return nil
		}
	case "QueryType":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.QueryType = string(tValue)
			return nil
		}
	case "UserAgent":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.UserAgent = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.UserID = string(tValue)
			return nil
		}
	case "Error":
		if tValue, ok := value.(lua.LString); ok {
			cnkr.Error = &servicelog.ErrorRecord{
				Name: string(tValue),
			}
			return nil
		}
	case "Args":
		if tValue, ok := value.(*lua.LTable); ok {
			cnkr.Args = scripting.LuaTableToMap(tValue)
			return nil
		}
	}
	return scripting.InvalidAttrError{Attr: name}
}
