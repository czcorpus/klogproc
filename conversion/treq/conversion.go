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

	"klogproc/conversion"
	"klogproc/logbuffer"
)

const (
	qTypeD = "D"
	qTypeL = "L"
)

// Transformer converts a Treq log record to a destination format
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {

	userID := -1
	if logRecord.UserID != "-" {
		uid, err := strconv.Atoi(logRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert user ID [%s]", logRecord.UserID)
		}
		userID = uid
	}

	isRegexp, err := conversion.ImportBool(logRecord.IsRegexp, "isRegexp")
	if err != nil {
		return nil, err
	}
	isCaseInsen, err := conversion.ImportBool(logRecord.IsCaseInsen, "isCaseInsen")
	if err != nil {
		return nil, err
	}
	isMultiWord, err := conversion.ImportBool(logRecord.IsMultiWord, "isMultiWord")
	if err != nil {
		return nil, err
	}
	isLemma, err := conversion.ImportBool(logRecord.IsMultiWord, "isLemma")
	if err != nil {
		return nil, err
	}

	out := &OutputRecord{
		Type:        "treq",
		time:        logRecord.GetTime(),
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		QLang:       logRecord.QLang,
		SecondLang:  logRecord.SecondLang,
		IPAddress:   logRecord.IPAddress,
		UserID:      logRecord.UserID,
		IsAnonymous: userID == -1 || conversion.UserBelongsToList(userID, anonymousUsers),
		// Corpus set later
		Subcorpus: logRecord.Subcorpus,
		// IsQuery set later
		IsRegexp:    isRegexp,
		IsCaseInsen: isCaseInsen,
		IsMultiWord: isMultiWord,
		IsLemma:     isLemma,
		QType:       logRecord.QType,
		Query:       logRecord.Query,
		Query2:      logRecord.Query2,
		// GeoIP set elsewhere
	}
	out.ID = createID(out)
	if out.QType == qTypeD {
		out.Corpus = fmt.Sprintf("intercorp_v8_%s", logRecord.QLang)
		out.IsQuery = true
	}
	return out, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return []conversion.InputRecord{rec}
}
