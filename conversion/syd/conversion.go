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

	"klogproc/conversion"
	"klogproc/logbuffer"
)

// Transformer converts a SyD log record to a destination format
type Transformer struct {
	Version     string
	SyncCorpora []string
	DiaCorpora  []string
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	var userID *int
	if logRecord.UserID != "-" {
		uid, err := strconv.Atoi(logRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert user ID [%s]", logRecord.UserID)
		}
		userID = &uid
	}

	r := &OutputRecord{
		Type:        recType,
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		time:        logRecord.GetTime(),
		IPAddress:   logRecord.IPAddress,
		UserID:      userID,
		IsAnonymous: userID == nil || conversion.UserBelongsToList(*userID, anonymousUsers),
		KeyReq:      logRecord.KeyReq,
		KeyUsed:     logRecord.KeyUsed,
		Key:         logRecord.Key,
		Ltool:       logRecord.Ltool,
		RunScript:   logRecord.RunScript,
		IsQuery:     true,
	}
	r.ID = createID(r)
	if logRecord.Ltool == "S" {
		r.Corpus = t.SyncCorpora

	} else if logRecord.Ltool == "D" {
		r.Corpus = t.DiaCorpora
	}
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return []conversion.InputRecord{rec}
}

// NewTransformer is a recommended factory for new Transformer instances
// to reflect the version properly
func NewTransformer(version string) *Transformer {
	switch version {
	case "0.1":
		return &Transformer{
			Version:     version,
			SyncCorpora: []string{"syn2010", "oral_v2", "ksk-dopisy"},
			DiaCorpora:  []string{"diakon"},
		}
	default:
		return nil
	}
}
