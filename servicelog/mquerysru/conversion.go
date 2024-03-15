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

package mquerysru

import (
	"klogproc/servicelog"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/rs/zerolog/log"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	ExcludeIPList []string
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	rec := &OutputRecord{
		Type:      recType,
		Datetime:  logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		datetime:  logRecord.GetTime(),
		Level:     logRecord.Level,
		IPAddress: logRecord.ClientIP,
		ProcTime:  logRecord.Latency,
		Error:     logRecord.ErrorMessage,

		Version:   logRecord.Version,
		Operation: logRecord.Operation,
		IsQuery:   logRecord.IsQuery(),

		Args: logRecord.Args,
	}
	rec.ID = CreateID(rec)
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if collections.SliceContains(t.ExcludeIPList, rec.GetClientIP().String()) {
		log.Debug().Str("ip", rec.GetClientIP().String()).Msg("excluded IP")
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}
