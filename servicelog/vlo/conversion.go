// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package vlo

import (
	"github.com/czcorpus/klogproc-core/storage"
	vloCore "github.com/czcorpus/klogproc-core/storage/vlo"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) AppType() string {
	return storage.AppTypeVLO
}

func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}
	rec := &vloCore.OutputRecord{
		Type:      t.AppType(),
		Level:     tLogRecord.Level,
		IPAddress: tLogRecord.ClientIP,
		ProcTime:  tLogRecord.Latency,
		Error:     tLogRecord.ExportError(),
		Operation: tLogRecord.Operation,
		IsQuery:   tLogRecord.IsQuery(),
	}
	rec.SetTime(tLogRecord.GetTime())
	rec.ID = rec.GenerateDeterministicID()
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return []storage.InputRecord{rec}, nil
}
