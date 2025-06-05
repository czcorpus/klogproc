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

import (
	"fmt"
	"strconv"
	"time"

	"klogproc/servicelog"
)

// Transformer converts a SyD log record to a destination format
type Transformer struct {
	version        string
	syncCorpora    []string
	diaCorpora     []string
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeSyd
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	var userID *int
	if tLogRecord.UserID != "-" {
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = &uid
	}

	r := &OutputRecord{
		Type:        t.AppType(),
		Datetime:    tLogRecord.GetTime().Format(time.RFC3339),
		time:        tLogRecord.GetTime(),
		IPAddress:   tLogRecord.IPAddress,
		UserID:      userID,
		IsAnonymous: userID == nil || servicelog.UserBelongsToList(*userID, t.anonymousUsers),
		KeyReq:      tLogRecord.KeyReq,
		KeyUsed:     tLogRecord.KeyUsed,
		Key:         tLogRecord.Key,
		Ltool:       tLogRecord.Ltool,
		RunScript:   tLogRecord.RunScript,
		IsQuery:     true,
	}
	r.ID = r.GenerateDeterministicID()
	if tLogRecord.Ltool == "S" {
		r.Corpus = t.syncCorpora

	} else if tLogRecord.Ltool == "D" {
		r.Corpus = t.diaCorpora
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

// NewTransformer is a recommended factory for new Transformer instances
// to reflect the version properly
func NewTransformer(
	version string,
	anonymousUsers []int,
) *Transformer {
	switch version {
	case "0.1":
		return &Transformer{
			version:        version,
			syncCorpora:    []string{"syn2010", "oral_v2", "ksk-dopisy"},
			diaCorpora:     []string{"diakon"},
			anonymousUsers: anonymousUsers,
		}
	default:
		return nil
	}
}
