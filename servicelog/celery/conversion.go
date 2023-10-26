// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package celery

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"klogproc/servicelog"
	"strconv"

	"github.com/rs/zerolog/log"
)

func createID(rec *OutputRecord) string {
	str := fmt.Sprintf("%v%v%v%v%v", rec.Clock, rec.Hostname, rec.ProcTime, rec.NumWorkerRestarts, rec.NumTaskCalls)
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// Transformer converts a source log object into a destination one
type Transformer struct{}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord) (*OutputRecord, error) {
	clk, err := strconv.Atoi(logRecord.Clock)
	if err != nil {
		log.Warn().Msgf("unable to get Clock value - %s", err)
	}

	out := &OutputRecord{
		Clock:        clk,
		Hostname:     logRecord.Broker.Hostname,
		ProcTime:     logRecord.Rusage.Stime + logRecord.Rusage.Utime,
		NumTaskCalls: make(map[string]int),
	}
	for k, v := range logRecord.Total {
		out.NumTaskCalls[k] = v
		out.NumTasksTotal += v
	}
	out.ID = createID(out)
	return out, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}
