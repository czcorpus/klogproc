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

package syd

// Transformer converts a SyD log record to a destination format
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string) (*OutputRecord, error) {
	var corpora []string
	if logRecord.Ltool == "S" {
		corpora = []string{"syn2010", "oral_v2", "ksk-dopisy"}

	} else if logRecord.Ltool == "D" {
		corpora = []string{"diakon"}
	}
	r := &OutputRecord{
		Type:      recType,
		Datetime:  logRecord.Datetime,
		IPAddress: logRecord.IPAddress,
		UserID:    logRecord.UserID,
		KeyReq:    logRecord.KeyReq,
		KeyUsed:   logRecord.KeyUsed,
		Key:       logRecord.Key,
		Ltool:     logRecord.Ltool,
		RunScript: logRecord.RunScript,
		IsQuery:   true,
		Corpus:    corpora,
	}
	r.ID = createID(r)
	return r, nil
}
