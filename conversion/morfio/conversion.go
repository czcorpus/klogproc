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

	"klogproc/conversion"
	"klogproc/logbuffer"
)

// Transformer converts a Morfio log record to a destination format
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {

	userID := -1
	if logRecord.UserID != "-" && logRecord.UserID != "" {
		uid, err := strconv.Atoi(logRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert user ID [%s]", logRecord.UserID)
		}
		userID = uid
	}

	minFreq, err := strconv.Atoi(logRecord.MinFreq)
	if err != nil {
		return nil, err
	}

	caseIns, err := conversion.ImportBool(logRecord.CaseInsensitive, "caseInsensitive")
	if err != nil {
		return nil, err
	}

	ans := &OutputRecord{
		// ID set later
		Type:            "morfio",
		time:            logRecord.GetTime(),
		Datetime:        logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:       logRecord.IPAddress,
		UserID:          logRecord.UserID,
		IsAnonymous:     userID == -1 || conversion.UserBelongsToList(userID, anonymousUsers),
		IsQuery:         true,
		KeyReq:          logRecord.KeyReq,
		KeyUsed:         logRecord.KeyUsed,
		Key:             logRecord.Key,
		RunScript:       logRecord.RunScript,
		Corpus:          logRecord.Corpus,
		MinFreq:         minFreq,
		InputAttr:       logRecord.InputAttr,
		OutputAttr:      logRecord.OutputAttr,
		CaseInsensitive: caseIns,
	}

	ans.ID = createID(ans)
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return []conversion.InputRecord{rec}
}
