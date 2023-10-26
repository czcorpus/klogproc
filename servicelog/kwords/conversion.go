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

package kwords

import (
	"fmt"
	"strconv"
	"time"

	"klogproc/servicelog"
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

	numFiles, err := strconv.Atoi(logRecord.NumFiles)
	if err != nil {
		return nil, err
	}
	targetLength, err := strconv.Atoi(logRecord.TargetLength)
	if err != nil {
		return nil, err
	}

	var refLength *int
	if logRecord.RefLength != "-" {
		rl, err := strconv.Atoi(logRecord.RefLength)
		if err != nil {
			return nil, err
		}
		refLength = &rl
	}
	pronouns, err := servicelog.ImportBool(logRecord.Prep, "pronouns")
	if err != nil {
		return nil, err
	}
	prep, err := servicelog.ImportBool(logRecord.Prep, "prep")
	if err != nil {
		return nil, err
	}
	con, err := servicelog.ImportBool(logRecord.Prep, "con")
	if err != nil {
		return nil, err
	}
	num, err := servicelog.ImportBool(logRecord.Prep, "num")
	if err != nil {
		return nil, err
	}
	caseInsen, err := servicelog.ImportBool(logRecord.Prep, "caseInsensitive")
	if err != nil {
		return nil, err
	}

	ans := &OutputRecord{
		// ID set later
		Type:            "kwords",
		time:            logRecord.GetTime(),
		Datetime:        logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:       logRecord.IPAddress,
		UserID:          logRecord.UserID,
		IsAnonymous:     userID == -1 || servicelog.UserBelongsToList(userID, anonymousUsers),
		IsQuery:         true,
		NumFiles:        numFiles,
		TargetInputType: logRecord.TargetInputType,
		TargetLength:    targetLength,
		Corpus:          logRecord.Corpus,
		RefLength:       refLength,
		Pronouns:        pronouns,
		Prep:            prep,
		Con:             con,
		Num:             num,
		CaseInsensitive: caseInsen,
		// GeoIP set elsewhere
	}

	ans.ID = createID(ans)
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}
