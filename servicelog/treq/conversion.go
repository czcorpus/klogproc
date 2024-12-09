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

package treq

import (
	"fmt"
	"strconv"
	"time"

	"klogproc/servicelog"
)

const (
	qTypeD = "D"
	qTypeL = "L"
)

// Transformer converts a Treq log record to a destination format
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeTreq
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
	if tLogRecord.UserID != "-" {
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = uid
	}

	isRegexp, err := servicelog.ImportBool(tLogRecord.IsRegexp, "isRegexp")
	if err != nil {
		return nil, err
	}
	isCaseInsen, err := servicelog.ImportBool(tLogRecord.IsCaseInsen, "isCaseInsen")
	if err != nil {
		return nil, err
	}
	isMultiWord, err := servicelog.ImportBool(tLogRecord.IsMultiWord, "isMultiWord")
	if err != nil {
		return nil, err
	}
	isLemma, err := servicelog.ImportBool(tLogRecord.IsMultiWord, "isLemma")
	if err != nil {
		return nil, err
	}

	out := &OutputRecord{
		Type:        "treq",
		time:        tLogRecord.GetTime(),
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		QLang:       tLogRecord.QLang,
		SecondLang:  tLogRecord.SecondLang,
		IPAddress:   tLogRecord.IPAddress,
		UserID:      tLogRecord.UserID,
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, t.AnonymousUsers),
		// Corpus set later
		Subcorpus: tLogRecord.Subcorpus,
		// IsQuery set later
		IsRegexp:    isRegexp,
		IsCaseInsen: isCaseInsen,
		IsMultiWord: isMultiWord,
		IsLemma:     isLemma,
		// GeoIP set elsewhere
	}
	out.ID = out.GenerateDeterministicID()
	if tLogRecord.QType == qTypeD {
		out.Corpus = fmt.Sprintf("intercorp_v8_%s", tLogRecord.QLang)
		out.IsQuery = true
	}
	return out, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}
