// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package kontext013

import (
	"strconv"
	"time"

	"klogproc/servicelog"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKontext
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	fullCorpname := importCorpname(tLogRecord)
	r := &OutputRecord{
		Type:           t.AppType(),
		Action:         tLogRecord.Action,
		Corpus:         fullCorpname.Corpname,
		AlignedCorpora: tLogRecord.GetAlignedCorpora(),
		Datetime:       tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:       tLogRecord.GetTime(),
		IPAddress:      tLogRecord.GetClientIP().String(),
		IsAnonymous:    servicelog.UserBelongsToList(tLogRecord.UserID, t.AnonymousUsers),
		IsQuery:        isEntryQuery(tLogRecord.Action),
		Limited:        fullCorpname.limited,
		ProcTime:       tLogRecord.ProcTime,
		QueryType:      importQueryType(tLogRecord),
		UserAgent:      tLogRecord.Request.HTTPUserAgent,
		UserID:         strconv.Itoa(tLogRecord.UserID),
		Error:          tLogRecord.Error.AsPointer(),
	}
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}
