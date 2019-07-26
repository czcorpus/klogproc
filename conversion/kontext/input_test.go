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

package kontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportDatetimeString(t *testing.T) {
	v, err := importDatetimeString("2019-07-25T23:13:57.3", "01:00")
	assert.Equal(t, "2019-07-25T23:13:57.3+01:00", v)
	assert.Nil(t, err)
}

// datetime where "T" is replaced by " "
func TestImportDatetimeString2(t *testing.T) {
	v, err := importDatetimeString("2019-07-25 23:13:57.3", "01:00")
	assert.Equal(t, "2019-07-25T23:13:57.3+01:00", v)
	assert.Nil(t, err)
}

// negative timezone
func TestImportDatetimeString3(t *testing.T) {
	v, err := importDatetimeString("2019-07-25 23:13:57.3", "-05:00")
	assert.Equal(t, "2019-07-25T23:13:57.3-05:00", v)
	assert.Nil(t, err)
}

func TestImportDatetimeInvalidString(t *testing.T) {
	v, err := importDatetimeString("foo", "00:00")
	assert.Equal(t, "", v)
	assert.Error(t, err)
}

func TestImportJSONLog(t *testing.T) {
	jsonLog := `{"user_id": 1537, "proc_time": 2.4023, "pid": 61800,
				"request": {"HTTP_X_FORWARDED_FOR": "89.176.43.98",
				"HTTP_USER_AGENT": "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0"},
				"action": "quick_filter", "params": {"ctxattrs": "word", "q2": "P0 1 1 [lemma=\"na\"]",
				"attr_vmode": "mouseover", "pagesize": "100", "refs": "=doc.title,=doc.pubyear", "q": "~Iz4HSjvL9mhP",
				"viewmode": "kwic", "attrs": "word", "corpname": "syn_v7", "attr_allpos": "all"}, "date": "2019-06-25 14:04:11.301121"}`

	rec, err := ImportJSONLog([]byte(jsonLog), "-01:00")
	assert.Nil(t, err)
	assert.Equal(t, 1537, rec.UserID)
	assert.InDelta(t, 2.4023, rec.ProcTime, 0.0001)
	assert.Equal(t, 61800, rec.PID)
	assert.Equal(t, "89.176.43.98", rec.Request.HTTPForwardedFor)
	assert.Equal(t, "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0", rec.Request.HTTPUserAgent)
	assert.Equal(t, "quick_filter", rec.Action)
	assert.Equal(t, "2019-06-25T14:04:11.301121-01:00", rec.Date)
	params := map[string]interface{}{
		"ctxattrs":    "word",
		"q2":          "P0 1 1 [lemma=\"na\"]",
		"attr_vmode":  "mouseover",
		"pagesize":    "100",
		"refs":        "=doc.title,=doc.pubyear",
		"q":           "~Iz4HSjvL9mhP",
		"viewmode":    "kwic",
		"attrs":       "word",
		"corpname":    "syn_v7",
		"attr_allpos": "all",
	}
	assert.Equal(t, params, rec.Params)
}

func TestImportJSONLogInvalid(t *testing.T) {
	jsonLog := `{}`
	rec, err := ImportJSONLog([]byte(jsonLog), "-01:00")
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestImportJSONLogDateOnly(t *testing.T) {
	jsonLog := `{"date": "2019-06-25 14:04:11.301121"}`
	rec, err := ImportJSONLog([]byte(jsonLog), "-01:00")
	assert.Nil(t, err)
	assert.Equal(t, "2019-06-25T14:04:11.301121-01:00", rec.Date)
}

func TestGetStringParam(t *testing.T) {
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Params: map[string]interface{}{
			"foo": 1,
			"bar": "xxx",
		},
	}
	assert.Equal(t, "", rec.GetStringParam("foo"))
	assert.Equal(t, "xxx", rec.GetStringParam("bar"))
	assert.Equal(t, "", rec.GetStringParam("baz"))
}

func TestGetIntParam(t *testing.T) {
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Params: map[string]interface{}{
			"foo": 1,
			"bar": "xxx",
		},
	}
	assert.Equal(t, 1, rec.GetIntParam("foo"))
	assert.Equal(t, -1, rec.GetIntParam("bar"))
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
		Date: "2019-06-25T14:04:50.23-01:00",
		Params: map[string]interface{}{
			"queryselector_intercorp_v10_cs": "value1",
			"queryselector_intercorp_v10_en": "value2",
			"pcq_pos_neg_intercorp_v10_de":   "value3",
		},
	}
	ac := rec.GetAlignedCorpora()
	assert.Contains(t, ac, "intercorp_v10_cs")
	assert.Contains(t, ac, "intercorp_v10_en")
	assert.Contains(t, ac, "intercorp_v10_de")
	assert.Equal(t, 3, len(ac))
}

func TestAgentIsBot(t *testing.T) {
	rec := InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: Request{
			HTTPUserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		},
	}
	assert.True(t, rec.AgentIsBot())
}
