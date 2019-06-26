// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

package record

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/czcorpus/klogproc/fetch"
)

// importQueryType translates KonText/Bonito query type argument
// into a more understandable form
func importQueryType(record *fetch.LogRecord) string {
	val := record.GetStringParam("queryselector")
	switch val {
	case "iqueryrow":
		return "basic"
	case "lemmarow":
		return "lemma"
	case "phraserow":
		return "phrase"
	case "wordrow":
		return "word"
	case "charrow":
		return "char"
	case "cqlrow":
		return "cql"
	default:
		return ""
	}
}

// importCorpname extracts actual corpus name from
// URL argument which may contain additional data (e.g. variant prefix)
func importCorpname(record *fetch.LogRecord) fullCorpname {
	var corpname string
	var limited bool

	if record.Params["corpname"] != "" {
		corpname = record.GetStringParam("corpname")
		corpname, _ = url.QueryUnescape(corpname)
		corpname = strings.Split(corpname, ";")[0]
		if strings.Index(corpname, "omezeni/") == 0 {
			corpname = corpname[len("omezeni/"):]
			limited = true

		} else {
			limited = false
		}
		return fullCorpname{corpname, limited}
	}
	return fullCorpname{}
}

// GeoDataRecord represents a full client geographical
// position information as provided by GeoIP database
type GeoDataRecord struct {
	ContinentCode string     `json:"continent_code"`
	CountryCode2  string     `json:"country_code2"`
	CountryCode3  string     `json:"country_code3"`
	CountryName   string     `json:"country_name"`
	IP            string     `json:"ip"`
	Latitude      float32    `json:"latitude"`
	Longitude     float32    `json:"longitude"`
	Location      [2]float32 `json:"location"`
	Timezone      string     `json:"timezone"`
}

// CNKRecord represents an exported application log record ready
// to be inserted into ElasticSearch index.
type CNKRecord struct {
	ID             string   `json:"-"`
	Type           string   `json:"-"`
	Action         string   `json:"action"`
	Corpus         string   `json:"corpus"`
	AlignedCorpora []string `json:"alignedCorpora"`
	Datetime       string   `json:"datetime"`
	datetime       time.Time
	IPAddress      string            `json:"ipAddress"`
	IsAnonymous    bool              `json:"isAnonymous"`
	IsQuery        bool              `json:"isQuery"`
	Limited        bool              `json:"limited"`
	ProcTime       float32           `json:"procTime"`
	QueryType      string            `json:"queryType"`
	Type2          string            `json:"type"` // TODO do we need this?
	UserAgent      string            `json:"userAgent"`
	UserID         int               `json:"userId"`
	GeoIP          GeoDataRecord     `json:"geoip"`
	Error          fetch.ErrorRecord `json:"error"`
}

// ToJSON converts self to JSON string
func (cnkr *CNKRecord) ToJSON() ([]byte, error) {
	return json.Marshal(cnkr)
}

// GetTime returns Go Time instance representing
// date and time when the record was created.
func (cnkr *CNKRecord) GetTime() time.Time {
	return cnkr.datetime
}

// New creates a new CNKRecord out of an existing LogRecord
func New(logRecord *fetch.LogRecord, recType string) *CNKRecord {
	fullCorpname := importCorpname(logRecord)
	r := &CNKRecord{
		Type:           recType,
		Action:         logRecord.Action,
		Corpus:         fullCorpname.Corpname,
		AlignedCorpora: logRecord.GetAlignedCorpora(),
		Datetime:       logRecord.Date,
		datetime:       logRecord.GetTime(),
		IPAddress:      logRecord.GetClientIP().String(),
		// IsAnonymous - not set here
		IsQuery:   isEntryQuery(logRecord.Action),
		Limited:   fullCorpname.limited,
		ProcTime:  logRecord.ProcTime,
		QueryType: importQueryType(logRecord),
		Type2:     recType,
		UserAgent: logRecord.Request.HTTPUserAgent,
		UserID:    logRecord.UserID,
		Error:     logRecord.Error,
	}
	r.ID = createID(r)
	return r
}

type fullCorpname struct {
	Corpname string
	limited  bool
}

func createID(cnkr *CNKRecord) string {
	str := cnkr.Action + cnkr.Corpus + cnkr.Datetime + cnkr.IPAddress +
		cnkr.Type + cnkr.UserAgent + strconv.Itoa(cnkr.UserID)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func isEntryQuery(action string) bool {
	ea := []string{"first", "wordlist", "wsketch", "thes", "wsdiff"}
	for _, item := range ea {
		if item == action {
			return true
		}
	}
	return false
}
