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

	"klogproc/servicelog"
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

func isInteractionPath(rec *InputRecord) bool {
	return rec.Path == "cunits/_view" || rec.Path == "/api/cunits/_view" || strings.HasPrefix(rec.Path, "/api/cunits/-const-acphrase")
}

func testIsQuery(rec *InputRecord) bool {
	return !testIsAPI(rec) && isInteractionPath(rec) && rec.Request.Page.From == 0
}

func testIsPhraseBank(rec *InputRecord) bool {
	return strings.HasPrefix(rec.Path, "/api/cunits/-const-acphrase")
}

// Transformer converts a KorpusDB log record to a destination format
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKorpusDB
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
	if tLogRecord.UserID != "" { // null is converted into an empty string
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = uid
	}

	out := &OutputRecord{
		Type:         t.AppType(),
		Datetime:     tLogRecord.GetTime().Format(time.RFC3339),
		time:         tLogRecord.GetTime(),
		Path:         tLogRecord.Path,
		Page:         tLogRecord.Request.Page,
		IPAddress:    tLogRecord.IP,
		UserID:       tLogRecord.UserID,
		ClientFlag:   tLogRecord.Request.ClientFlag,
		IsAnonymous:  userID == -1 || servicelog.UserBelongsToList(userID, t.AnonymousUsers),
		IsQuery:      testIsQuery(tLogRecord),
		IsAPI:        testIsAPI(tLogRecord),
		IsPhraseBank: testIsPhraseBank(tLogRecord),
		QueryType:    getQueryType(tLogRecord),
	}
	out.ID = out.GenerateDeterministicID()
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

func NewTransformer() *Transformer {
	return &Transformer{}
}
