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

	"klogproc/conversion"
	"klogproc/conversion/wag06"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*wag06.OutputRecord, error) {
	rec := wag06.NewTimedOutputRecord(logRecord.GetTime(), tzShiftMin)
	rec.Type = recType
	rec.Action = logRecord.Action
	rec.IPAddress = logRecord.Request.Origin
	rec.UserAgent = logRecord.Request.HTTPUserAgent
	rec.ReferringDomain = logRecord.Request.Referer
	rec.UserID = strconv.Itoa(logRecord.UserID)
	rec.IsAnonymous = conversion.UserBelongsToList(logRecord.UserID, anonymousUsers)
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

func (t *Transformer) HistoryLookupSecs() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs []conversion.InputRecord,
) conversion.InputRecord {
	return rec
}
