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
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
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
	case "user/ajax_query_history":
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
	analyzer      *analysis.BotAnalyzer[*QueryInputRecord]
	ExcludeIPList servicelog.ExcludeIPList
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *QueryInputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	corpname := importCorpname(logRecord)
	r := &OutputRecord{
		Type:           recType,
		Action:         logRecord.Action,
		Corpus:         corpname,
		AlignedCorpora: logRecord.GetAlignedCorpora(),
		Datetime:       logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:       logRecord.GetTime(),
		IPAddress:      logRecord.GetClientIP().String(),
		IsAnonymous:    servicelog.UserBelongsToList(logRecord.UserID, anonymousUsers),
		IsQuery:        isEntryQuery(logRecord.Action) && !logRecord.IsIndirectCall,
		ProcTime:       logRecord.ProcTime,
		QueryType:      importQueryType(logRecord),
		UserAgent:      logRecord.Request.HTTPUserAgent,
		UserID:         strconv.Itoa(logRecord.UserID),
		Error:          logRecord.Error.AsPointer(),
		Args:           exportArgs(logRecord.Action, logRecord.Args),
	}
	r.ID = createID(r)
	return r, nil
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
	bufferConf *load.BufferConf,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
	excludeIPList []string,
) *Transformer {
	analyzer := analysis.NewBotAnalyzer[*QueryInputRecord]("kontext", bufferConf, realtimeClock, emailNotifier)
	return &Transformer{
		analyzer:      analyzer,
		ExcludeIPList: excludeIPList,
	}
}
