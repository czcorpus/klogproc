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

package morfio

import (
	"fmt"
	"strconv"
	"time"

	"klogproc/servicelog"
)

// Transformer converts a Morfio log record to a destination format
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeMorfio
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	userID := -1
	if tLogRecord.UserID != "-" && tLogRecord.UserID != "" {
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = uid
	}

	minFreq, err := strconv.Atoi(tLogRecord.MinFreq)
	if err != nil {
		return nil, err
	}

	caseIns, err := servicelog.ImportBool(tLogRecord.CaseInsensitive, "caseInsensitive")
	if err != nil {
		return nil, err
	}

	ans := &OutputRecord{
		// ID set later
		Type:            "morfio",
		time:            tLogRecord.GetTime(),
		Datetime:        tLogRecord.GetTime().Format(time.RFC3339),
		IPAddress:       tLogRecord.IPAddress,
		UserID:          tLogRecord.UserID,
		IsAnonymous:     userID == -1 || servicelog.UserBelongsToList(userID, t.AnonymousUsers),
		IsQuery:         true,
		KeyReq:          tLogRecord.KeyReq,
		KeyUsed:         tLogRecord.KeyUsed,
		Key:             tLogRecord.Key,
		RunScript:       tLogRecord.RunScript,
		Corpus:          tLogRecord.Corpus,
		MinFreq:         minFreq,
		InputAttr:       tLogRecord.InputAttr,
		OutputAttr:      tLogRecord.OutputAttr,
		CaseInsensitive: caseIns,
	}

	ans.ID = ans.GenerateDeterministicID()
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}
