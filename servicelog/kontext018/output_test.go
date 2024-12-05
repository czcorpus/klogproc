// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package kontext018

import (
	"klogproc/servicelog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// cnkr.Action + cnkr.Corpus + cnkr.Datetime + cnkr.IPAddress + cnkr.Type + cnkr.UserAgent + cnkr.UserID
func createRecord() *OutputRecord {
	return &OutputRecord{
		ID:          "abcdef",
		Type:        servicelog.AppTypeKontext,
		Action:      "view",
		Corpus:      "syn2015",
		Datetime:    "2017-02-11T11:02:31.880",
		IPAddress:   "195.113.53.66",
		IsAnonymous: true,
		IsQuery:     false,
		ProcTime:    0.712,
		QueryType:   "cql",
		UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:51.0) Gecko/20100101 Firefox/51.0",
		UserID:      "100",
		GeoIP: servicelog.GeoDataRecord{
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
	if rec.GenerateDeterministicID() != "2452d6c39ddd4dfcba2df61e1115511e547c09af" {
		t.Error("Hash match error")
	}
}

func TestImportCorpname(t *testing.T) {
	p := make(map[string]interface{})
	p["corpname"] = "foobar7"
	r := &QueryInputRecord{Args: p}
	c := importCorpname(r)
	assert.Equal(t, "foobar7", c)
}
