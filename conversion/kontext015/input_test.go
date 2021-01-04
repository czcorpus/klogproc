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

package kontext015

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportDatetimeString(t *testing.T) {
	v, err := importDatetimeString("2019-07-25T23:13:57.3")
	assert.Equal(t, "2019-07-25T23:13:57.3", v)
	assert.Nil(t, err)
}

// datetime where "T" is replaced by " "
func TestImportDatetimeString2(t *testing.T) {
	v, err := importDatetimeString("2019-07-25 23:13:57.3")
	assert.Equal(t, "2019-07-25T23:13:57.3", v)
	assert.Nil(t, err)
}

// negative timezone
func TestImportDatetimeStringWithTimezoneToIgnore(t *testing.T) {
	v, err := importDatetimeString("2019-07-25 23:13:57.3+02:00")
	assert.Equal(t, "2019-07-25T23:13:57.3", v)
	assert.Nil(t, err)
}

func TestImportDatetimeStringWithSuffixZ(t *testing.T) {
	v, err := importDatetimeString("2019-07-25T23:13:57.3Z")
	assert.Equal(t, "2019-07-25T23:13:57.3", v)
	assert.Nil(t, err)
}

func TestImportDatetimeInvalidString(t *testing.T) {
	v, err := importDatetimeString("foo")
	assert.Equal(t, "", v)
	assert.Error(t, err)
}

func TestImportJSONLog(t *testing.T) {
	jsonLog := `{"args": {"corpora": ["ortofon_v1", "ortofox_v1"], "qtype": "simple", "use_regexp": false,
	"qmcase": false, "extended_query": false, "uses_context": 0, "uses_tt": false},
	"date": "2021-01-03 01:28:46.153734", "action": "query_submit", "user_id": 19185,
	"proc_time": 0.5779, "request": {"HTTP_X_FORWARDED_FOR": "89.176.43.98",
	"HTTP_USER_AGENT": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0"}}`

	rec, err := ImportJSONLog([]byte(jsonLog))
	assert.Nil(t, err)
	assert.Equal(t, 19185, rec.UserID)
	assert.InDelta(t, 0.5779, rec.ProcTime, 0.0001)
	assert.Equal(t, "89.176.43.98", rec.Request.HTTPForwardedFor)
	assert.Equal(t, "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0", rec.Request.HTTPUserAgent)
	assert.Equal(t, "query_submit", rec.Action)
	assert.Equal(t, "2021-01-03T01:28:46.153734", rec.Date)
	args := map[string]interface{}{
		"corpora":        []interface{}{"ortofon_v1", "ortofox_v1"},
		"qtype":          "simple",
		"use_regexp":     false,
		"qmcase":         false,
		"extended_query": false,
		"uses_context":   false,
		"uses_tt":        false,
	}
	assert.Equal(t, args, rec.Args)
}

func TestImportJSONLogInvalid(t *testing.T) {
	jsonLog := `{}`
	rec, err := ImportJSONLog([]byte(jsonLog))
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestImportJSONLogDateOnly(t *testing.T) {
	jsonLog := `{"date": "2019-06-25 14:04:11.301121"}`
	rec, err := ImportJSONLog([]byte(jsonLog))
	assert.Nil(t, err)
	assert.Equal(t, "2019-06-25T14:04:11.301121", rec.Date)
}

func TestGetStringArg(t *testing.T) {
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Args: map[string]interface{}{
			"foo": map[string]interface{}{
				"zoo": "hit",
			},
			"bar": "xxx",
		},
	}
	assert.Equal(t, "", rec.GetStringArg("xoo"))
	assert.Equal(t, "", rec.GetStringArg("foo", "baz"))
	assert.Equal(t, "xxx", rec.GetStringArg("bar"))
	assert.Equal(t, "hit", rec.GetStringArg("foo", "zoo"))
}

func TestGetIntParam(t *testing.T) {
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Args: map[string]interface{}{
			"foo": 1,
			"bar": "xxx",
		},
	}
	assert.Equal(t, 1, rec.GetIntArg("foo"))
	assert.Equal(t, -1, rec.GetIntArg("bar"))
}

func TestGetClientSimple(t *testing.T) {
	remoteIP := "89.176.43.98"
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: Request{
			HTTPRemoteAddr: remoteIP,
		},
	}
	assert.Equal(t, remoteIP, rec.GetClientIP().String())
}

func TestGetClientSimple2(t *testing.T) {
	remoteIP := "89.176.43.98"
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: Request{
			RemoteAddr: remoteIP,
		},
	}
	assert.Equal(t, remoteIP, rec.GetClientIP().String())
}

func TestGetClientIPProxy(t *testing.T) {
	forwIP := "89.176.43.98"
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: Request{
			HTTPForwardedFor: forwIP,
			HTTPRemoteAddr:   "127.0.0.1",
			RemoteAddr:       "127.0.0.1",
		},
	}
	assert.Equal(t, forwIP, rec.GetClientIP().String())
}

func TestGetAlignedCorpora(t *testing.T) {
	rec := InputRecord{
		Action: "query_submit",
		Date:   "2019-06-25T14:04:50.23-01:00",
		Args: map[string]interface{}{
			"corpora": []interface{}{"intercorp_v11_cs", "intercorp_v11_en", "intercorp_v11_de"},
		},
	}
	ac := rec.GetAlignedCorpora()
	assert.Contains(t, ac, "intercorp_v11_en")
	assert.Contains(t, ac, "intercorp_v11_de")
	assert.Equal(t, 2, len(ac))
}

func TestGetAlignedCorporaNonQueryAction(t *testing.T) {
	rec := InputRecord{
		Action: "wordlist",
		Date:   "2019-06-25T14:04:50.23-01:00",
		Args: map[string]interface{}{
			"corpora": []interface{}{"intercorp_v11_cs", "intercorp_v11_en", "intercorp_v11_de"},
		},
	}
	ac := rec.GetAlignedCorpora()
	assert.Equal(t, 0, len(ac))
}
