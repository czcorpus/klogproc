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

	"github.com/czcorpus/klogproc-core/storage"
	shinyCore "github.com/czcorpus/klogproc-core/storage/shiny"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	appType        string
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return t.appType
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	userID := tLogRecord.User.ID
	if userID == 0 && len(t.anonymousUsers) > 0 {
		userID = t.anonymousUsers[0]
	}
	ans := &shinyCore.OutputRecord{
		Type:        t.AppType(),
		IsQuery:     true,
		IPAddress:   tLogRecord.ClientIP,
		User:        tLogRecord.User.User,
		UserID:      strconv.Itoa(userID),
		IsAnonymous: storage.UserBelongsToList(userID, t.anonymousUsers),
		Lang:        tLogRecord.Lang,
		UserAgent:   tLogRecord.UserAgent,
	}
	ans.SetTime(tLogRecord.GetTime())
	ans.ID = ans.GenerateDeterministicID()
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return []storage.InputRecord{rec}, nil
}

func NewTransformer(
	appType string,
	anonymousUsers []int,
) *Transformer {
	if appType != storage.AppTypeAkalex && appType != storage.AppTypeCalc &&
		appType != storage.AppTypeGramatikat && appType != storage.AppTypeQuitaUp &&
		appType != storage.AppTypeLists {
		panic("invalid application type for a Shiny transformer")
	}
	return &Transformer{
		appType:        appType,
		anonymousUsers: anonymousUsers,
	}
}
