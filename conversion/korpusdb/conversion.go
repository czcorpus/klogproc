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

package korpusdb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"klogproc/conversion"
	"klogproc/logbuffer"
)

func getQueryType(rec *InputRecord) string {
	switch rec.Request.Query.Type {
	case ":token:form":
		return "token"
	case ":ngram:form:*":
		return "ngram"
	default:
		return ""
	}
}

func testIsAPI(rec *InputRecord) bool {
	return rec.Request.ClientFlag != "" && !strings.HasPrefix(rec.Request.ClientFlag, "ratatosk-paw/")
}

func testIsQuery(rec *InputRecord) bool {
	return !testIsAPI(rec) && rec.Path == "cunits/_view" && rec.Request.Page.From == 0
}

// Transformer converts a KorpusDB log record to a destination format
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {

	userID := -1
	if logRecord.UserID != "" { // null is converted into an empty string
		uid, err := strconv.Atoi(logRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", logRecord.UserID)
		}
		userID = uid
	}

	out := &OutputRecord{
		Type:        conversion.AppTypeKorpusDB,
		time:        logRecord.GetTime(),
		Path:        logRecord.Path,
		Page:        logRecord.Request.Page,
		IPAddress:   logRecord.IP,
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		UserID:      logRecord.UserID,
		ClientFlag:  logRecord.Request.ClientFlag,
		IsAnonymous: userID == -1 || conversion.UserBelongsToList(userID, anonymousUsers),
		IsQuery:     testIsQuery(logRecord),
		IsAPI:       testIsAPI(logRecord),
		QueryType:   getQueryType(logRecord),
	}
	out.ID = createID(out)
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
