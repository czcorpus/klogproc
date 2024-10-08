// Copyright 2021 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2021 Institute of the Czech National Corpus,
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

package wag07

import (
	"strconv"

	"klogproc/analysis"
	"klogproc/load"
	"klogproc/notifications"
	"klogproc/servicelog"
	"klogproc/servicelog/wag06"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	analyzer      servicelog.Preprocessor
	excludeIPList servicelog.ExcludeIPList
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*wag06.OutputRecord, error) {
	rec := wag06.NewTimedOutputRecord(logRecord.GetTime(), tzShiftMin)
	rec.Type = recType
	rec.Action = logRecord.Action
	rec.IPAddress = logRecord.Request.Origin
	rec.UserAgent = logRecord.Request.HTTPUserAgent
	rec.ReferringDomain = logRecord.Request.Referer
	rec.UserID = strconv.Itoa(logRecord.UserID)
	rec.IsAnonymous = servicelog.UserBelongsToList(logRecord.UserID, anonymousUsers)
	rec.IsQuery = logRecord.IsQuery
	rec.IsMobileClient = logRecord.IsMobileClient
	rec.HasPosSpecification = logRecord.HasPosSpecification
	rec.QueryType = logRecord.QueryType
	rec.Lang1 = logRecord.Lang1
	rec.Lang2 = logRecord.Lang2
	rec.Queries = []string{} // no more used?
	rec.ProcTime = -1        // TODO not available; does it have a value
	rec.ID = wag06.CreateID(rec)
	return rec, nil
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
	return t.analyzer.Preprocess(rec, prevRecs)
}

func NewTransformer(
	bufferConf *load.BufferConf,
	excludeIPList []string,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
) *Transformer {
	var analyzer servicelog.Preprocessor
	if bufferConf != nil && bufferConf.BotDetection != nil {
		analyzer = analysis.NewBotAnalyzer[*InputRecord]("wag", bufferConf, realtimeClock, emailNotifier)

	} else {
		analyzer = analysis.NewNullAnalyzer[*InputRecord]("wag")
	}
	return &Transformer{
		analyzer:      analyzer,
		excludeIPList: excludeIPList,
	}
}
