// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package kontext015

import (
	"strconv"
	"time"

	"klogproc/scripting"
	"klogproc/servicelog"

	lua "github.com/yuin/gopher-lua"
)

func exportArgs(data map[string]interface{}) map[string]interface{} {
	ans := make(map[string]interface{})
	for k, v := range data {
		if k != "corpora" && k != "corpname" {
			ans[k] = v
		}
	}
	return ans
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	ExcludeIPList  servicelog.ExcludeIPList
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKontext
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
	corpname := importCorpname(tLogRecord)
	r := &OutputRecord{
		Type:           t.AppType(),
		Action:         tLogRecord.Action,
		Corpus:         corpname,
		AlignedCorpora: tLogRecord.GetAlignedCorpora(),
		Datetime:       tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:       tLogRecord.GetTime(),
		IPAddress:      tLogRecord.GetClientIP().String(),
		IsAnonymous:    servicelog.UserBelongsToList(tLogRecord.UserID, t.AnonymousUsers),
		IsQuery:        isEntryQuery(tLogRecord.Action) && !tLogRecord.IsIndirectCall,
		ProcTime:       tLogRecord.ProcTime,
		QueryType:      importQueryType(tLogRecord),
		UserAgent:      tLogRecord.Request.HTTPUserAgent,
		UserID:         strconv.Itoa(tLogRecord.UserID),
		Error:          tLogRecord.Error.AsPointer(),
		Args:           exportArgs(tLogRecord.Args),
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
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
