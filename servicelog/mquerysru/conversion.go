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

package mquerysru

import (
	"klogproc/servicelog"
	"strings"
	"time"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	ExcludeIPList servicelog.ExcludeIPList
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeMquerySRU
}

func (t *Transformer) getCorpus(logRecord *InputRecord) string {
	corpus := logRecord.Args.XFCSContext
	if corpus != "" {
		return t.corpusPID2ID(corpus)

	} else {
		if len(logRecord.Args.Sources) > 0 {
			return strings.Join(logRecord.Args.Sources, "+")
		}
	}
	return ""
}

func (t *Transformer) corpusPID2ID(s string) string {
	tmp := strings.Split(s, "/")
	return tmp[len(tmp)-1]
}

func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	rec := &OutputRecord{
		Type:      t.AppType(),
		Datetime:  tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:  tLogRecord.GetTime(),
		Level:     tLogRecord.Level,
		IPAddress: tLogRecord.ClientIP,
		ProcTime:  tLogRecord.Latency,
		Error:     tLogRecord.ExportError(),
		Corpus:    t.getCorpus(tLogRecord),
		Version:   tLogRecord.Version,
		Operation: tLogRecord.Operation,
		IsQuery:   tLogRecord.IsQuery(),
		Args:      tLogRecord.Args,
	}
	rec.ID = rec.GenerateDeterministicID()
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
