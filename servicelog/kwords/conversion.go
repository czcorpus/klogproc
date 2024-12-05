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
type Transformer struct {
	ExcludeIPList  servicelog.ExcludeIPList
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKwords
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
	if tLogRecord.UserID != "-" && tLogRecord.UserID != "" {
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = uid
	}

	numFiles, err := strconv.Atoi(tLogRecord.NumFiles)
	if err != nil {
		return nil, err
	}
	targetLength, err := strconv.Atoi(tLogRecord.TargetLength)
	if err != nil {
		return nil, err
	}

	var refLength *int
	if tLogRecord.RefLength != "-" {
		rl, err := strconv.Atoi(tLogRecord.RefLength)
		if err != nil {
			return nil, err
		}
		refLength = &rl
	}
	pronouns, err := servicelog.ImportBool(tLogRecord.Prep, "pronouns")
	if err != nil {
		return nil, err
	}
	prep, err := servicelog.ImportBool(tLogRecord.Prep, "prep")
	if err != nil {
		return nil, err
	}
	con, err := servicelog.ImportBool(tLogRecord.Prep, "con")
	if err != nil {
		return nil, err
	}
	num, err := servicelog.ImportBool(tLogRecord.Prep, "num")
	if err != nil {
		return nil, err
	}
	caseInsen, err := servicelog.ImportBool(tLogRecord.Prep, "caseInsensitive")
	if err != nil {
		return nil, err
	}

	ans := &OutputRecord{
		// ID set later
		Type:            "kwords",
		time:            tLogRecord.GetTime(),
		Datetime:        tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:       tLogRecord.IPAddress,
		UserID:          tLogRecord.UserID,
		IsAnonymous:     userID == -1 || servicelog.UserBelongsToList(userID, t.AnonymousUsers),
		IsQuery:         true,
		NumFiles:        numFiles,
		TargetInputType: tLogRecord.TargetInputType,
		TargetLength:    targetLength,
		Corpus:          tLogRecord.Corpus,
		RefLength:       refLength,
		Pronouns:        pronouns,
		Prep:            prep,
		Con:             con,
		Num:             num,
		CaseInsensitive: caseInsen,
		// GeoIP set elsewhere
	}

	ans.ID = ans.GenerateDeterministicID()
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if t.ExcludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}
