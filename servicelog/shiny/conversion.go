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

package shiny

import (
	"strconv"
	"time"

	"klogproc/scripting"
	"klogproc/servicelog"

	lua "github.com/yuin/gopher-lua"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	excludeIPList  servicelog.ExcludeIPList
	appType        string
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return t.appType
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
	userID := tLogRecord.User.ID
	if userID == 0 && len(t.anonymousUsers) > 0 {
		userID = t.anonymousUsers[0]
	}
	ans := &OutputRecord{
		Type:        t.AppType(),
		time:        tLogRecord.GetTime(),
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IsQuery:     true,
		IPAddress:   tLogRecord.ClientIP,
		User:        tLogRecord.User.User,
		UserID:      strconv.Itoa(userID),
		IsAnonymous: servicelog.UserBelongsToList(userID, t.anonymousUsers),
		Lang:        tLogRecord.Lang,
		UserAgent:   tLogRecord.UserAgent,
	}
	ans.ID = createID(ans)
	return ans, nil
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

func NewTransformer(
	appType string,
	excludeIPList servicelog.ExcludeIPList,
	anonymousUsers []int,
) *Transformer {
	if appType != servicelog.AppTypeAkalex && appType != servicelog.AppTypeCalc &&
		appType != servicelog.AppTypeGramatikat && appType != servicelog.AppTypeQuitaUp &&
		appType != servicelog.AppTypeLists {
		panic("invalid application type for a Shiny transformer")
	}
	return &Transformer{
		appType:        appType,
		excludeIPList:  excludeIPList,
		anonymousUsers: anonymousUsers,
	}
}
