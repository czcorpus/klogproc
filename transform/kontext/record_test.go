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

package kontext

import (
	"fmt"
	"testing"

	"github.com/czcorpus/klogproc/fetch"
)

// cnkr.Action + cnkr.Corpus + cnkr.Datetime + cnkr.IPAddress + cnkr.Type + cnkr.UserAgent + cnkr.UserID
func createRecord() *CNKRecord {
	return &CNKRecord{
		ID:          "abcdef",
		Type:        "kontext",
		Action:      "view",
		Corpus:      "syn2015",
		Datetime:    "2017-02-11T11:02:31.880",
		IPAddress:   "195.113.53.66",
		IsAnonymous: true,
		IsQuery:     false,
		Limited:     false,
		ProcTime:    0.712,
		QueryType:   "cql",
		Type2:       "kontext",
		UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:51.0) Gecko/20100101 Firefox/51.0",
		UserID:      100,
		GeoIP: GeoDataRecord{
			CountryCode2: "CZ",
			CountryName:  "Czech Republic",
			IP:           "195.113.53.66",
			Latitude:     49.4,
			Longitude:    17.674,
			Location:     [2]float32{17.6742, 49.3996},
		},
	}
}

func TestCreateID(t *testing.T) {
	rec := createRecord()
	fmt.Println("HASH: ", createID(rec))
	if createID(rec) != "2452d6c39ddd4dfcba2df61e1115511e547c09af" {
		t.Error("Hash match error")
	}
}

func TestImportCorpname(t *testing.T) {
	p := make(map[string]interface{})
	p["corpname"] = "omezeni/foobar7;x=10"
	r := &fetch.LogRecord{Params: p}
	c := importCorpname(r)
	if c.Corpname != "foobar7" || c.limited != true {
		t.Error("Failed import corpname: ", c)
	}
}
