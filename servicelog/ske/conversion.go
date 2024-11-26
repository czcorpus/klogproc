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

	"klogproc/scripting"
	"klogproc/servicelog"
	"klogproc/users"

	lua "github.com/yuin/gopher-lua"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	userMap        *users.UserMap
	excludeIPList  servicelog.ExcludeIPList
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeSke
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
	userID := -1
	if tLogRecord.User != "-" && tLogRecord.User != "" {
		uid := t.userMap.GetIdOf(tLogRecord.User)
		if uid < 0 {
			return nil, fmt.Errorf("failed to find user ID of [%s]", tLogRecord.User)
		}
		userID = uid
	}

	corpname, isLimited := importCorpname(tLogRecord.Corpus)
	r := &OutputRecord{
		Type:        t.AppType(),
		time:        tLogRecord.GetTime(),
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:   tLogRecord.Request.RemoteAddr,
		UserAgent:   tLogRecord.Request.HTTPUserAgent,
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, t.anonymousUsers),
		IsQuery:     isEntryQuery(tLogRecord.Action),
		UserID:      strconv.Itoa(userID),
		Action:      tLogRecord.Action,
		Corpus:      corpname,
		Limited:     isLimited,
		Subcorpus:   tLogRecord.Subcorpus,
		ProcTime:    tLogRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if t.excludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(
	userMap *users.UserMap,
	excludeIPList servicelog.ExcludeIPList,
	anonymousUsers []int,
) *Transformer {
	return &Transformer{
		userMap:        userMap,
		excludeIPList:  excludeIPList,
		anonymousUsers: anonymousUsers,
	}
}
