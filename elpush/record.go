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

package elpush

import (
	"crypto/sha1"
	"encoding/hex"
	// "fmt"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/czcorpus/klogproc/logs"
)

func importQueryType(record *logs.LogRecord) string {
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

type CNKRecordMeta struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
	Type  string `json:"_type"`
}

type ElasticCNKRecordMeta struct {
	Index CNKRecordMeta `json:"index"`
}

func (ecrm *ElasticCNKRecordMeta) ToJSON() ([]byte, error) {
	return json.Marshal(ecrm)
}

type CNKRecord struct {
	ID          string        `json:"-"`
	Type        string        `json:"-"`
	Action      string        `json:"action"`
	Corpus      string        `json:"corpus"`
	Datetime    string        `json:"datetime"`
	IPAddress   string        `json:"ipAddress"`
	IsAnonymous bool          `json:"isAnonymous"`
	IsQuery     bool          `json:"isQuery"`
	Limited     bool          `json:"limited"`
	ProcTime    float32       `json:"procTime"`
	QueryType   string        `json:"queryType"`
	Type2       string        `json:"type"` // TODO do we need this?
	UserAgent   string        `json:"userAgent"`
	UserID      int           `json:"userId"`
	GeoIP       GeoDataRecord `json:"geoip"`
}

func (cnkr *CNKRecord) ToJSON() ([]byte, error) {
	return json.Marshal(cnkr)
}

// TODO do we need this?
type CNKRecordError struct {
	Message string
	Cause   error
}

func (cnkre *CNKRecordError) Error() string {
	return fmt.Sprintf("CNKRecordError: %s", cnkre.Message)
}

// New creates a new CNKRecord out of an existing LogRecord
func New(logRecord *logs.LogRecord, recType string) *CNKRecord {
	fullCorpname := importCorpname(logRecord)
	r := &CNKRecord{
		Type:      recType,
		Action:    logRecord.Action,
		Corpus:    fullCorpname.Corpname,
		Datetime:  logRecord.Date,
		IPAddress: logRecord.GetClientIP().String(),
		// IsAnonymous - not set here
		IsQuery:   isEntryQuery(logRecord.Action),
		Limited:   fullCorpname.limited,
		ProcTime:  logRecord.ProcTime,
		QueryType: importQueryType(logRecord),
		Type2:     recType,
		UserAgent: logRecord.Request.HTTPUserAgent,
		UserID:    logRecord.UserID,
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

func importCorpname(record *logs.LogRecord) fullCorpname {
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
