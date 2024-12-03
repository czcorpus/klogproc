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

	"klogproc/scripting"
	"klogproc/servicelog"

	lua "github.com/yuin/gopher-lua"
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
type Transformer struct {
	ExcludeIPList  servicelog.ExcludeIPList
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKorpusDB
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
	if tLogRecord.UserID != "" { // null is converted into an empty string
		uid, err := strconv.Atoi(tLogRecord.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user ID [%s]", tLogRecord.UserID)
		}
		userID = uid
	}

	out := &OutputRecord{
		Type:        t.AppType(),
		Datetime:    tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		time:        tLogRecord.GetTime(),
		Path:        tLogRecord.Path,
		Page:        tLogRecord.Request.Page,
		IPAddress:   tLogRecord.IP,
		UserID:      tLogRecord.UserID,
		ClientFlag:  tLogRecord.Request.ClientFlag,
		IsAnonymous: userID == -1 || servicelog.UserBelongsToList(userID, t.AnonymousUsers),
		IsQuery:     testIsQuery(tLogRecord),
		IsAPI:       testIsAPI(tLogRecord),
		QueryType:   getQueryType(tLogRecord),
	}
	out.ID = createID(out)
	return out, nil
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	tRec, ok := rec.(*OutputRecord)
	if !ok {
		return scripting.ErrFailedTypeAssertion
	}
	switch name {
	case "Type":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Type = string(tValue)
			return nil
		}
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			tRec.time = servicelog.ConvertDatetimeString(string(tValue))
			tRec.Datetime = string(tValue)
			return nil
		}
	case "Path":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Path = string(tValue)
			return nil
		}
	case "Page":
		if tValue, ok := value.(*lua.LTable); ok {
			fromVal := tValue.RawGetString("From")
			if tFromVal, ok := fromVal.(lua.LNumber); ok {
				tRec.Page.From = int(tFromVal)
			}
			sizeVal := tValue.RawGetString("Size")
			if tSizeVal, ok := sizeVal.(lua.LNumber); ok {
				tRec.Page.Size = int(tSizeVal)
			}
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			tRec.IPAddress = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			tRec.UserID = string(tValue)
			return nil
		}
	case "ClientFlag":
		if tValue, ok := value.(lua.LString); ok {
			tRec.ClientFlag = string(tValue)
			return nil
		}
	case "IsAnonymous":
		tRec.IsAnonymous = value == lua.LTrue
		return nil
	case "IsQuery":
		tRec.IsQuery = value == lua.LTrue
		return nil
	case "QueryType":
		if tValue, ok := value.(lua.LString); ok {
			tRec.QueryType = string(tValue)
			return nil
		}
	}
	return scripting.InvalidAttrError{Attr: name}
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

func NewTransformer(
	excludeIPList servicelog.ExcludeIPList,
) *Transformer {
	return &Transformer{
		ExcludeIPList: excludeIPList,
	}
}
