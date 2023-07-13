// Copyright 2022 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2022 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2022 Institute of the Czech National Corpus,
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

package masm

import (
	"klogproc/conversion"
	"klogproc/logbuffer"
	"time"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	rec := &OutputRecord{
		time:           logRecord.GetTime(),
		Datetime:       logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		Type:           recType,
		Level:          logRecord.Level,
		Message:        logRecord.Message,
		IsQuery:        logRecord.IsQuery,
		Corpus:         logRecord.Corpus,
		AlignedCorpora: logRecord.AlignedCorpora,
		IsAutocomplete: logRecord.IsAutocomplete,
		IsCached:       logRecord.IsCached,
		ProcTimeSecs:   logRecord.ProcTimeSecs,
		Error:          logRecord.Error,
	}
	rec.ID = CreateID(rec)
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return []conversion.InputRecord{rec}
}
