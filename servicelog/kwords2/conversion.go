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
	"fmt"
	"klogproc/servicelog"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	textSplitPattern = regexp.MustCompile(`\s+`)
)

func convertMultitypeInt(v any) int {
	switch tv := v.(type) {
	case int:
		return tv
	case float64:
		return int(math.Round(tv))
	case string:
		v, err := strconv.Atoi(tv)
		if err != nil {
			return -1
		}
		return v
	}
	return -1
}

func getNumOfWords(text string) int {
	tmp := strings.TrimSpace(text)
	if len(tmp) == 0 {
		return 0
	}
	return len(textSplitPattern.Split(tmp, -1))
}

// --

type Transformer struct{}

func (t *Transformer) getActionName(rec *InputRecord) string {
	items := strings.Split(rec.Path, "/")
	if len(items) > 0 {
		return items[len(items)-1]
	}
	return ""
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}

func (t *Transformer) Transform(
	logRecord *InputRecord,
	recType string,
	tzShiftMin int,
	anonymousUsers []int,
) (*OutputRecord, error) {
	aName := t.getActionName(logRecord)
	var args *Args
	if aName == "keywords" {
		args = &Args{
			Attrs:        logRecord.Body.Attrs,
			Level:        logRecord.Body.Level,
			EffectMetric: logRecord.Body.EffectMetric,
			MinFreq:      convertMultitypeInt(logRecord.Body.MinFreq),
			Percent:      convertMultitypeInt(logRecord.Body.Percent),
		}
	}
	r := &OutputRecord{
		Type:              recType,
		Action:            t.getActionName(logRecord),
		Corpus:            logRecord.Body.RefCorpus,
		NumInputTextWords: getNumOfWords(logRecord.Body.Text),
		TextLang:          logRecord.Body.Lang,
		Datetime:          logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:         logRecord.GetClientIP().String(),
		IsAnonymous:       true, // TODO !!!
		IsQuery:           t.getActionName(logRecord) == "keywords",
		UserAgent:         logRecord.Headers.UserAgent,
		UserID:            fmt.Sprint(logRecord.UserID),
		Error:             logRecord.Exception,
		Args:              args,
		Version:           "2",
	}
	r.ID = createID(r)
	return r, nil
}
