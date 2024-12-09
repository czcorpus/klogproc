// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package apiguard

import (
	"klogproc/servicelog"
	"strconv"
	"time"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeAPIGuard
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
	var sUserID string
	if tLogRecord.UserID != nil {
		sUserID = strconv.Itoa(*tLogRecord.UserID)
	}
	corrDT := logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin))
	r := &OutputRecord{
		Type:       tLogRecord.Type,
		IsQuery:    true,
		Service:    tLogRecord.Service,
		ProcTime:   tLogRecord.ProcTime,
		IsCached:   tLogRecord.IsCached,
		IsIndirect: tLogRecord.IsIndirect,
		UserID:     sUserID,
		IPAddress:  tLogRecord.IPAddress,
		UserAgent:  tLogRecord.UserAgent,
		datetime:   corrDT,
		Datetime:   corrDT.Format(time.RFC3339),
	}
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}
