// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package mapka2

import (
	"strconv"
	"time"

	"klogproc/servicelog"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeMapka
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

	r := &OutputRecord{
		Type:        t.AppType(),
		time:        tLogRecord.GetTime(),
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:   tLogRecord.Request.RemoteAddr,
		UserAgent:   tLogRecord.Request.HTTPUserAgent,
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, t.anonymousUsers),
		IsQuery:     false,
		UserID:      strconv.Itoa(userID),
		Action:      tLogRecord.Action,
		Path:        tLogRecord.Path,
		ProcTime:    tLogRecord.ProcTime,
	}
	r.ID = r.GenerateDeterministicID()
	if r.Action == "index" || r.Action == "records_list" || r.Action == "city" {
		r.IsQuery = true
	}
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(
	anonymousUsers []int,
) *Transformer {
	return &Transformer{
		anonymousUsers: anonymousUsers,
	}
}
