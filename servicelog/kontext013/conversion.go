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

	"github.com/czcorpus/klogproc-core/storage"
	kontext013Core "github.com/czcorpus/klogproc-core/storage/kontext013"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeKontext
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	fullCorpname := importCorpname(tLogRecord)
	r := &kontext013Core.OutputRecord{
		Type:           t.AppType(),
		Action:         tLogRecord.Action,
		Corpus:         fullCorpname.Corpname,
		AlignedCorpora: tLogRecord.GetAlignedCorpora(),
		IPAddress:      tLogRecord.GetClientIP().String(),
		IsAnonymous:    storage.UserBelongsToList(tLogRecord.UserID, t.AnonymousUsers),
		IsQuery:        kontext013Core.IsEntryQuery(tLogRecord.Action),
		Limited:        fullCorpname.limited,
		ProcTime:       tLogRecord.ProcTime,
		QueryType:      importQueryType(tLogRecord),
		UserAgent:      tLogRecord.Request.HTTPUserAgent,
		UserID:         strconv.Itoa(tLogRecord.UserID),
		Error:          tLogRecord.Error.AsPointer(),
	}
	r.SetTime(tLogRecord.GetTime())
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return []storage.InputRecord{rec}, nil
}
