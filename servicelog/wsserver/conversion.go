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
	"klogproc/servicelog"
	"strings"
)

func cleanIPInfo(ip string) string {
	return strings.Split(ip, ",")[0]
}

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeWsserver
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	ans := &OutputRecord{
		Action:    tLogRecord.Action,
		Corpus:    tLogRecord.Corpus,
		Model:     tLogRecord.Model,
		time:      tLogRecord.GetTime(),
		ProcTime:  tLogRecord.ProcTime,
		IsQuery:   true,
		IPAddress: cleanIPInfo(tLogRecord.IPAddress),
		UserAgent: tLogRecord.HTTPUserAgent,
		UserID:    "-1",
	}

	ans.ID = ans.GenerateDeterministicID()
	return ans, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}
