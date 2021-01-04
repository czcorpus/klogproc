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

	"github.com/czcorpus/klogproc/conversion"
	"github.com/stretchr/testify/assert"
)

type onErrorHandler struct {
	Msg string
}

func (h *onErrorHandler) OnError(message string) {
	h.Msg = message
}

func (h *onErrorHandler) Evaluate() {}

func (h *onErrorHandler) Reset() {}

func TestParseLine(t *testing.T) {
	line := `2018-03-06 19:34:40,755 [QUERY] INFO: {"user_id": 4230, "proc_time": 0.5398, "pid": 46885, "request": {"HTTP_X_FORWARDED_FOR": "66.249.65.216", "HTTP_USER_AGENT": "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"}, "action": "view", "params": {"ctxattrs": "word", "attr_vmode": "visible", "pagesize": "50", "q": "~U1OTWBoC", "viewmode": "kwic", "attrs": "word", "corpname": "omezeni/syn2015", "structs": "p", "attr_allpos": "kw"}, "date": "2018-03-06 19:34:40.755029"}`
	p := LineParser013{}
	rec, err := p.ParseLine(line, 71)
	assert.Nil(t, err)
	assert.NotNil(t, rec)
}

func TestParseLineInvalidQuery(t *testing.T) {
	line := `2018-03-06 19:34:40,755 [QUERY] INFO: this won't work`
	p := LineParser013{}
	rec, err := p.ParseLine(line, 71)
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestParseLineNonQueryLine(t *testing.T) {
	line := `2018-03-06 22:07:57,836 [actions.concordance] ERROR: AttrNotFound (lemma)`
	p := LineParser013{appErrorRegister: &onErrorHandler{}}
	rec, err := p.ParseLine(line, 71)
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestGetLineType(t *testing.T) {
	line := `2018-03-06 22:07:57,836 [actions.concordance] ERROR: AttrNotFound (lemma)`
	tp := getLineType(line)
	assert.Equal(t, "ERROR", tp)

	line = `2018-03-06 19:34:40,755 [QUERY] INFO: {"date": "2018-03-06 19:34:40.755029"}`
	tp = getLineType(line)
	assert.Equal(t, "INFO", tp)

	line = `2018-03-06 19:34:40,755 [QUERY] WARNING: no configuration available`
	tp = getLineType(line)
	assert.Equal(t, "WARNING", tp)
}

func TestIsIgnoredError(t *testing.T) {
	lines := []string{
		"2020-03-19 09:22:10,484 [actions.concordance] ERROR: regexopt: at position 2: error, unexpected LBRACKET",
		"2020-03-23 08:30:25,097 [actions.concordance] ERROR: syntax error, unexpected $end near position 3",
	}
	for _, line := range lines {
		reg := &onErrorHandler{}
		p := LineParser013{appErrorRegister: reg}
		rec, err := p.ParseLine(line, 71)
		_, ok := err.(conversion.LineParsingError)
		assert.True(t, ok)
		assert.Nil(t, rec)
		assert.Equal(t, "", reg.Msg)
	}
}

func TestIsNotIgnoredError(t *testing.T) {

	lines := []string{
		"2020-03-19 09:22:10,484 [actions.concordance] ERROR: ValueError",
		"2020-03-23 08:31:02,382 [controller] ERROR: need more than 1 value to unpack",
		"2020-03-23 08:30:46,350 [actions.concordance] ERROR: AttrNotFound (word and 1=1)",
	}
	for _, line := range lines {
		reg := &onErrorHandler{}
		p := LineParser013{appErrorRegister: reg}
		rec, err := p.ParseLine(line, 71)
		_, ok := err.(conversion.LineParsingError)
		assert.True(t, ok)
		assert.Nil(t, rec)
		assert.Equal(t, line, reg.Msg)
	}
}
