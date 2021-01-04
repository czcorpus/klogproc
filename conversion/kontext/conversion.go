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

package kontext

import (
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

// Transformer converts a source log object into a destination one
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	fullCorpname := importCorpname(logRecord)
	r := &OutputRecord{
		Type:           recType,
		Action:         logRecord.Action,
		Corpus:         fullCorpname.Corpname,
		AlignedCorpora: logRecord.GetAlignedCorpora(),
		Datetime:       logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:       logRecord.GetTime(),
		IPAddress:      logRecord.GetClientIP().String(),
		IsAnonymous:    conversion.UserBelongsToList(logRecord.UserID, anonymousUsers),
		IsQuery:        isEntryQuery(logRecord.Action),
		Limited:        fullCorpname.limited,
		ProcTime:       logRecord.ProcTime,
		QueryType:      importQueryType(logRecord),
		UserAgent:      logRecord.Request.HTTPUserAgent,
		UserID:         logRecord.UserID,
		Error:          logRecord.Error,
	}
	r.ID = createID(r)
	return r, nil
}
