// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string) (*OutputRecord, error) {
	fullCorpname := importCorpname(logRecord)
	r := &OutputRecord{
		Type:           recType,
		Action:         logRecord.Action,
		Corpus:         fullCorpname.Corpname,
		AlignedCorpora: logRecord.GetAlignedCorpora(),
		Datetime:       logRecord.Date,
		datetime:       logRecord.GetTime(),
		IPAddress:      logRecord.GetClientIP().String(),
		// IsAnonymous - not set here
		IsQuery:   isEntryQuery(logRecord.Action),
		Limited:   fullCorpname.limited,
		ProcTime:  logRecord.ProcTime,
		QueryType: importQueryType(logRecord),
		Type2:     recType,
		UserAgent: logRecord.Request.HTTPUserAgent,
		UserID:    logRecord.UserID,
		Error:     logRecord.Error,
	}
	r.ID = createID(r)
	return r, nil
}
