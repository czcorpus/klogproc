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

package wag

import (
	"time"
)

func isQuery(action string) bool {
	return action == actionSearch || action == actionCompare || action == actionTranslate
}

// Transformer converts a source log object into a destination one
type Transformer struct {
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, anonymousUsers []int) (*OutputRecord, error) {

	r := &OutputRecord{
		Type:                recType,
		time:                logRecord.GetTime(),
		Datetime:            logRecord.GetTime().Format(time.RFC3339),
		IPAddress:           logRecord.Request.RemoteAddr,
		UserAgent:           logRecord.Request.HTTPUserAgent,
		IsAnonymous:         true, // currently we
		IsQuery:             isQuery(logRecord.Action),
		HasPosSpecification: logRecord.HasPosSpecification,
		QueryType:           logRecord.QueryType,
		Lang1:               logRecord.Lang1,
		Lang2:               logRecord.Lang2,
		Queries:             logRecord.Queries,
		Action:              logRecord.Action,
		ProcTime:            logRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}
