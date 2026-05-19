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

	"github.com/czcorpus/klogproc-core/analysis"
	"github.com/czcorpus/klogproc-core/logbuffer"
	"github.com/czcorpus/klogproc-core/storage"
	wag06Core "github.com/czcorpus/klogproc-core/storage/wag06"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	analyzer       storage.Preprocessor
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeWag
}

func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	rec := wag06Core.NewTimedOutputRecord(tLogRecord.GetTime())
	rec.Type = t.AppType()
	rec.Action = tLogRecord.Action
	rec.IPAddress = tLogRecord.Request.Origin
	rec.UserAgent = tLogRecord.Request.HTTPUserAgent
	rec.ReferringDomain = tLogRecord.Request.Referer
	rec.UserID = strconv.Itoa(tLogRecord.UserID)
	rec.IsAnonymous = storage.UserBelongsToList(tLogRecord.UserID, t.anonymousUsers)
	rec.IsQuery = tLogRecord.IsQuery
	rec.IsMobileClient = tLogRecord.IsMobileClient
	rec.HasPosSpecification = tLogRecord.HasPosSpecification
	rec.QueryType = tLogRecord.QueryType
	rec.Lang1 = tLogRecord.Lang1
	rec.Lang2 = tLogRecord.Lang2
	rec.Queries = []string{} // no more used?
	rec.ProcTime = -1        // TODO not available; does it have a value
	rec.ID = rec.GenerateDeterministicID()
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return t.analyzer.Preprocess(rec, prevRecs), nil
}

func NewTransformer(
	bufferConf *logbuffer.BufferConf,
	anonymousUsers []int,
	realtimeClock bool,
	emailNotifier analysis.Notifier,
) *Transformer {
	var analyzer storage.Preprocessor
	if bufferConf != nil && bufferConf.BotDetection != nil {
		analyzer = analysis.NewBotAnalyzer[*InputRecord]("wag", bufferConf, realtimeClock, emailNotifier)

	} else {
		analyzer = analysis.NewNullAnalyzer[*InputRecord]("wag")
	}
	return &Transformer{
		analyzer:       analyzer,
		anonymousUsers: anonymousUsers,
	}
}
