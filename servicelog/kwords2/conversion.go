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
	"klogproc/servicelog"
	"strings"
	"time"
)

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
	r := &OutputRecord{
		Type:        recType,
		Action:      t.getActionName(logRecord),
		Corpus:      logRecord.Body.RefCorpus,
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:   logRecord.GetClientIP().String(),
		IsAnonymous: true, // TODO !!!
		IsQuery:     t.getActionName(logRecord) == "keywords",
		UserAgent:   logRecord.Headers.UserAgent,
		UserID:      "0", // TODO !!!
		Error:       logRecord.Exception,
		Args: Args{
			Attrs:        logRecord.Body.Attrs,
			Level:        logRecord.Body.Level,
			EffectMetric: logRecord.Body.EffectMetric,
			MinFreq:      logRecord.Body.MinFreq,
			Percent:      logRecord.Body.Percent,
		},
	}
	r.ID = createID(r)
	return r, nil
}
