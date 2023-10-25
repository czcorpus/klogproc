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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"klogproc/logbuffer"
	"klogproc/servicelog"
	"strconv"
	"time"
)

func createID(apgr *OutputRecord) string {
	str := apgr.Datetime + strconv.FormatBool(apgr.IsQuery) + apgr.Service + apgr.Type +
		apgr.IPAddress + apgr.UserAgent + fmt.Sprintf("%01.3f", apgr.ProcTime) +
		strconv.FormatBool(apgr.IsCached) + strconv.FormatBool(apgr.IsIndirect)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// Transformer converts a source log object into a destination one
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord *InputRecord,
	recType string,
	tzShiftMin int,
	anonymousUsers []int,
) (*OutputRecord, error) {
	var sUserID string
	if logRecord.UserID != nil {
		sUserID = strconv.Itoa(*logRecord.UserID)
	}
	corrDT := logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin))
	r := &OutputRecord{
		Type:       logRecord.Type,
		IsQuery:    true,
		Service:    logRecord.Service,
		ProcTime:   logRecord.ProcTime,
		IsCached:   logRecord.IsCached,
		IsIndirect: logRecord.IsIndirect,
		UserID:     sUserID,
		IPAddress:  logRecord.IPAddress,
		UserAgent:  logRecord.UserAgent,
		datetime:   corrDT,
		Datetime:   corrDT.Format(time.RFC3339),
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}
