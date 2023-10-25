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
	"klogproc/servicelog"
	"net/url"
	"time"
)

// importQueryType translates KonText/Bonito query type argument
// into a more understandable form
func importQueryType(record *InputRecord) string {
	return record.GetStringArg("qtype")
}

// importCorpname extracts actual corpus name from
// URL argument which may contain additional data (e.g. variant prefix)
func importCorpname(record *InputRecord) string {
	var corpname string

	if record.HasArg("corpname") && record.GetStringArg("corpname") != "" {
		corpname = record.GetStringArg("corpname")
		corpname, _ = url.QueryUnescape(corpname)
		return corpname

	} else if record.HasArg("corpora") {
		c, ok := getSliceOfStrings(record.Args, "corpora")
		if ok && len(c) > 0 {
			return c[0]
		}
	}
	return ""
}

// OutputRecord represents an exported application log record ready
// to be inserted into ElasticSearch index.
type OutputRecord struct {
	ID             string   `json:"-"`
	Type           string   `json:"type"`
	Action         string   `json:"action"`
	Corpus         string   `json:"corpus"`
	AlignedCorpora []string `json:"alignedCorpora"`
	Datetime       string   `json:"datetime"`
	datetime       time.Time
	IPAddress      string                   `json:"ipAddress"`
	IsAnonymous    bool                     `json:"isAnonymous"`
	IsQuery        bool                     `json:"isQuery"`
	ProcTime       float32                  `json:"procTime"`
	QueryType      string                   `json:"queryType"`
	UserAgent      string                   `json:"userAgent"`
	UserID         string                   `json:"userId"`
	GeoIP          servicelog.GeoDataRecord `json:"geoip,omitempty"`
	Error          ErrorRecord              `json:"error"`
	Args           map[string]interface{}   `json:"args"`
}

// ToJSON converts self to JSON string
func (cnkr *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(cnkr)
}

func (cnkr *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	tags = make(map[string]string)
	values = make(map[string]interface{})
	values["procTime"] = cnkr.ProcTime
	values["error"] = cnkr.Error.Name
	values["errorAnchor"] = cnkr.Error.Anchor
	tags["corpname"] = cnkr.Corpus
	tags["queryType"] = cnkr.QueryType
	tags["action"] = cnkr.Action
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

func createID(cnkr *OutputRecord) string {
	str := cnkr.Action + cnkr.Corpus + cnkr.Datetime + cnkr.IPAddress +
		cnkr.Type + cnkr.UserAgent + cnkr.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func isEntryQuery(action string) bool {
	ea := []string{"/query_submit", "/wordlist/submit", "/pquery/freq_intersection"}
	for _, item := range ea {
		if item == action {
			return true
		}
	}
	return false
}
