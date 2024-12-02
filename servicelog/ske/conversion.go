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
		Corpus:      corpname,
		Subcorpus:   tLogRecord.Subcorpus,
		Limited:     isLimited,
		Action:      tLogRecord.Action,
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		time:        tLogRecord.GetTime(),
		IPAddress:   tLogRecord.Request.RemoteAddr,
		UserAgent:   tLogRecord.Request.HTTPUserAgent,
		UserID:      strconv.Itoa(userID),
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, t.anonymousUsers),
		IsQuery:     isEntryQuery(tLogRecord.Action),
		ProcTime:    tLogRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	tRec, ok := rec.(*OutputRecord)
	if !ok {
		return scripting.ErrFailedTypeAssertion
	}
	switch name {
	case "Type":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Type = string(tValue)
			return nil
		}
	case "Corpus":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Corpus = string(tValue)
			return nil
		}
	case "Subcorpus":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Subcorpus = string(tValue)
			return nil
		}
	case "Limited":
		if tValue, ok := value.(lua.LBool); ok {
			if tValue == lua.LTrue {
				tRec.Limited = true
			}
			return nil
		}
	case "Action":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Action = string(tValue)
			return nil
		}
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Datetime = string(tValue)
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			tRec.IPAddress = string(tValue)
			return nil
		}
	case "UserAgent":
		if tValue, ok := value.(lua.LString); ok {
			tRec.UserAgent = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			tRec.UserID = string(tValue)
			return nil
		}
	case "IsAnonymous":
		tRec.IsAnonymous = value == lua.LTrue
		return nil
	case "IsQuery":
		tRec.IsQuery = value == lua.LTrue
		return nil
	case "ProcTime":
		if tValue, ok := value.(lua.LNumber); ok {
			tRec.ProcTime = float64(tValue)
			return nil
		}
	}
	return scripting.InvalidAttrError{Attr: name}
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
// Note: Due to the fact that SkE is no more used in the CNC,
// UserMap loading was removed from Klogproc. In case it will
// be needed in the future, a custom Lua script should be used
// for this (users.UserMap has a Lua interface).
func NewTransformer(
	excludeIPList servicelog.ExcludeIPList,
	anonymousUsers []int,
) *Transformer {
	return &Transformer{
		userMap:        users.EmptyUserMap(),
		excludeIPList:  excludeIPList,
		anonymousUsers: anonymousUsers,
	}
}
