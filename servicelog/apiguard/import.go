// Copyright 2025 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2025 Institute of the Czech National Corpus,
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
	"klogproc/servicelog/mquery"
	"strings"
)

func ImportAPIGuardRecordAsMQuery(logRecord *InputRecord) *mquery.OutputRecord {
	var action, corpusID string
	split := strings.Split(logRecord.Path, "/")
	if len(split) == 2 {
		action = split[1]
	} else if len(split) >= 3 {
		action = split[len(split)-2]
		corpusID = split[len(split)-1]
	}
	r := &mquery.OutputRecord{
		Type:      servicelog.AppTypeMquery,
		Level:     logRecord.Level,
		IPAddress: logRecord.IPAddress,
		UserAgent: logRecord.GetUserAgent(),
		IsAI:      strings.Contains(logRecord.GetUserAgent(), "GPT"),
		ProcTime:  logRecord.ProcTime,
		Action:    action,
		CorpusID:  corpusID,
	}
	r.SetTime(logRecord.GetTime())
	return r
}
