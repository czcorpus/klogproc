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

package ske

import (
	"fmt"
	"strconv"
	"time"

	"klogproc/servicelog"
	"klogproc/users"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	userMap *users.UserMap
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	userID := -1
	if logRecord.User != "-" && logRecord.User != "" {
		uid := t.userMap.GetIdOf(logRecord.User)
		if uid < 0 {
			return nil, fmt.Errorf("Failed to find user ID of [%s]", logRecord.User)
		}
		userID = uid
	}

	corpname, isLimited := importCorpname(logRecord.Corpus)
	r := &OutputRecord{
		Type:        recType,
		time:        logRecord.GetTime(),
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:   logRecord.Request.RemoteAddr,
		UserAgent:   logRecord.Request.HTTPUserAgent,
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, anonymousUsers),
		IsQuery:     isEntryQuery(logRecord.Action),
		UserID:      strconv.Itoa(userID),
		Action:      logRecord.Action,
		Corpus:      corpname,
		Limited:     isLimited,
		Subcorpus:   logRecord.Subcorpus,
		ProcTime:    logRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(userMap *users.UserMap) *Transformer {
	return &Transformer{userMap: userMap}
}
