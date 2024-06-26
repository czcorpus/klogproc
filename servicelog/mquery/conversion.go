// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package mquery

import (
	"klogproc/servicelog"
	"time"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	ExcludeIPList servicelog.ExcludeIPList
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	rec := &OutputRecord{
		Type:      recType,
		Datetime:  logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:  logRecord.GetTime(),
		Level:     logRecord.Level,
		IPAddress: logRecord.ClientIP,
		UserAgent: logRecord.GetUserAgent(),
		IsAI:      logRecord.IsAI(),
		ProcTime:  logRecord.Latency,
		Error:     logRecord.ExportError(),
		Action:    logRecord.GetAction(),
		CorpusID:  logRecord.CorpusId,
	}
	rec.ID = CreateID(rec)
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if t.ExcludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}
