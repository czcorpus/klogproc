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

	"klogproc/users"

	"github.com/czcorpus/klogproc-core/storage"
	skeCore "github.com/czcorpus/klogproc-core/storage/ske"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	userMap        *users.UserMap
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeSke
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	userID := -1
	if tLogRecord.User != "-" && tLogRecord.User != "" {
		uid := t.userMap.GetIdOf(tLogRecord.User)
		if uid < 0 {
			return nil, fmt.Errorf("failed to find user ID of [%s]", tLogRecord.User)
		}
		userID = uid
	}

	corpname, isLimited := skeCore.ImportCorpname(tLogRecord.Corpus)
	r := &skeCore.OutputRecord{
		Type:        t.AppType(),
		Corpus:      corpname,
		Subcorpus:   tLogRecord.Subcorpus,
		Limited:     isLimited,
		Action:      tLogRecord.Action,
		IPAddress:   tLogRecord.Request.RemoteAddr,
		UserAgent:   tLogRecord.Request.HTTPUserAgent,
		UserID:      strconv.Itoa(userID),
		IsAnonymous: userID == -1 || storage.UserBelongsToList(userID, t.anonymousUsers),
		IsQuery:     skeCore.IsEntryQuery(tLogRecord.Action),
		ProcTime:    tLogRecord.ProcTime,
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

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
// Note: Due to the fact that SkE is no more used in the CNC,
// UserMap loading was removed from Klogproc. In case it will
// be needed in the future, a custom Lua script should be used
// for this (users.UserMap has a Lua interface).
func NewTransformer(
	anonymousUsers []int,
) *Transformer {
	return &Transformer{
		userMap:        users.EmptyUserMap(),
		anonymousUsers: anonymousUsers,
	}
}
