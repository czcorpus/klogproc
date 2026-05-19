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

package mapka3

import (
	"github.com/czcorpus/klogproc-core/analysis/clustering"
	"github.com/czcorpus/klogproc-core/logbuffer"
	"github.com/czcorpus/klogproc-core/storage"
	mapka3Core "github.com/czcorpus/klogproc-core/storage/mapka3"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	bufferConf     *logbuffer.BufferConf
	analyzer       storage.Preprocessor
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeMapka
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	r := &mapka3Core.OutputRecord{
		Type:      t.AppType(),
		IPAddress: tLogRecord.GetClientIP().String(),
		UserAgent: tLogRecord.GetUserAgent(),
		IsAnonymous: tLogRecord.Extra.UserID == "" ||
			storage.UserBelongsToList(tLogRecord.Extra.UserID, t.anonymousUsers),
		Action:      "interaction",
		UserID:      tLogRecord.Extra.UserID,
		ClusterSize: tLogRecord.clusterSize,
	}
	if r.ClusterSize > 0 {
		r.IsQuery = true
	}
	r.SetTime(tLogRecord.GetTime())
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return t.bufferConf.HistoryLookupItems
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord,
	prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return t.analyzer.Preprocess(rec, prevRecs), nil
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(
	bufferConf *logbuffer.BufferConf,
	anonymousUsers []int,
	realtimeClock bool,
) *Transformer {
	return &Transformer{
		bufferConf:     bufferConf,
		anonymousUsers: anonymousUsers,
		analyzer:       clustering.NewAnalyzer[*InputRecord]("mapka", bufferConf, realtimeClock),
	}
}
