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

package wsserver

import (
	"crypto/sha1"
	"encoding/hex"
	"klogproc/servicelog"
	"strings"
	"time"
)

// createID creates an idempotent ID of rec based on its properties.
func createID(rec *OutputRecord) string {
	str := rec.Type + rec.Action + rec.GetTime().Format(time.RFC3339) + rec.IPAddress + rec.UserID +
		rec.Action + rec.Model + rec.Corpus
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

func cleanIPInfo(ip string) string {
	return strings.Split(ip, ",")[0]
}

// Transformer converts a source log object into a destination one
type Transformer struct {
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	ans := &OutputRecord{
		Action:    logRecord.Action,
		Corpus:    logRecord.Corpus,
		Model:     logRecord.Model,
		time:      logRecord.GetTime(),
		ProcTime:  logRecord.ProcTime,
		IsQuery:   true,
		IPAddress: cleanIPInfo(logRecord.IPAddress),
		UserAgent: logRecord.HTTPUserAgent,
		UserID:    "-1",
	}

	ans.ID = createID(ans)
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}
