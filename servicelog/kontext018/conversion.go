// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package kontext018

import (
	"reflect"
	"strconv"
	"time"

	"klogproc/analysis"
	"klogproc/load"
	"klogproc/notifications"
	"klogproc/scripting"
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

func convertUrlValue(v string, tryBool bool) any {
	ans, err := strconv.Atoi(v)
	if err != nil {
		return v
	}
	if tryBool {
		return ans > 0
	}
	return ans
}

func exportArgs(action string, data map[string]interface{}) map[string]interface{} {
	ans := make(map[string]interface{})
	switch action {
	case "/user/ajax_query_history":
		for k, v := range data {
			switch tv := v.(type) {
			case string:
				switch k {
				case "extended_search":
					ans[k] = convertUrlValue(tv, true)
				default:
					ans[k] = convertUrlValue(tv, false)
				}
			case []string:
				ans[k] = v
			default:
				log.Error().
					Str("attr", k).
					Str("foundType", reflect.TypeOf(v).String()).
					Msg("kontext18 conversion expects `args` to contain only string or []string values")
			}
		}
	default:
		for k, v := range data {
			if k != "corpora" && k != "corpname" {
				ans[k] = v
			}
		}
	}
	return ans
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	analyzer       *analysis.BotAnalyzer[*QueryInputRecord]
	excludeIPList  servicelog.ExcludeIPList
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKontext
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*QueryInputRecord)
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
		IsAnonymous:    servicelog.UserBelongsToList(tLogRecord.UserID, t.anonymousUsers),
		IsQuery:        isEntryQuery(tLogRecord.Action) && !tLogRecord.IsIndirectCall,
		ProcTime:       tLogRecord.ProcTime,
		QueryType:      importQueryType(tLogRecord),
		UserAgent:      tLogRecord.Request.HTTPUserAgent,
		UserID:         strconv.Itoa(tLogRecord.UserID),
		Error:          tLogRecord.Error.AsPointer(),
		Args:           exportArgs(tLogRecord.Action, tLogRecord.Args),
	}
	r.ID = createID(r)
	return r, nil
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
	case "Action":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Action = string(tValue)
			return nil
		}
	case "Corpus":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Corpus = string(tValue)
			return nil
		}
	case "AlignedCorpora":
		if tValue, ok := value.(*lua.LTable); ok {
			var err error
			tRec.AlignedCorpora, err = scripting.LuaTableToSliceOfStrings(tValue)
			return err
		}
	case "Datetime":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Datetime = string(tValue)
			return nil
		}
	case "IPAddress":
		if tValue, ok := value.(lua.LString); ok {
			tRec.IPAddress = string(tValue)
			return nil
		}
	case "IsAnonymous":
		tRec.IsAnonymous = value == lua.LTrue
		return nil
	case "IsQuery":
		tRec.IsQuery = value == lua.LTrue
		return nil
	case "ProcTime":
		if tValue, ok := value.(lua.LNumber); ok {
			tRec.ProcTime = float64(tValue)
			return nil
		}
	case "QueryType":
		if tValue, ok := value.(lua.LString); ok {
			tRec.QueryType = string(tValue)
			return nil
		}
	case "UserAgent":
		if tValue, ok := value.(lua.LString); ok {
			tRec.UserAgent = string(tValue)
			return nil
		}
	case "UserID":
		if tValue, ok := value.(lua.LString); ok {
			tRec.UserID = string(tValue)
			return nil
		}
	case "Error":
		if tValue, ok := value.(lua.LString); ok {
			tRec.Error = &servicelog.ErrorRecord{
				Name: string(tValue),
			}
			return nil
		}
	case "Args":
		if tValue, ok := value.(*lua.LTable); ok {
			tRec.Args = scripting.LuaTableToMap(tValue)
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
	if t.excludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}

func NewTransformer(
	bufferConf *load.BufferConf,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
	excludeIPList []string,
	anonymousUsers []int,
) *Transformer {
	analyzer := analysis.NewBotAnalyzer[*QueryInputRecord]("kontext", bufferConf, realtimeClock, emailNotifier)
	return &Transformer{
		analyzer:       analyzer,
		excludeIPList:  excludeIPList,
		anonymousUsers: anonymousUsers,
	}
}
