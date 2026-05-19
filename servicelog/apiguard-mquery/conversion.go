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

package apiguardMquery

import (
	"klogproc/servicelog/apiguard"
	"strings"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/klogproc-core/storage"
	apiguardMqueryCore "github.com/czcorpus/klogproc-core/storage/mquery"
	"github.com/rs/zerolog/log"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) AppType() string {
	return storage.AppTypeAPIGuardMquery
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*apiguard.InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}

	var action, corpusID string
	split := strings.Split(tLogRecord.RequestPath, "/")
	if len(split) == 2 {
		action = split[1]

	} else if len(split) >= 3 {
		action = split[len(split)-2]
		corpusID = split[len(split)-1]
	}
	r := &apiguardMqueryCore.OutputRecord{
		Type:      storage.AppTypeMquery,
		Level:     tLogRecord.Level,
		IsCached:  tLogRecord.IsCached,
		IPAddress: tLogRecord.IPAddress,
		UserAgent: tLogRecord.GetUserAgent(),
		IsAI:      unireq.IsAIBot(tLogRecord),
		ProcTime:  tLogRecord.ProcTime,
		Action:    action,
		CorpusID:  corpusID,
	}
	r.SetTime(logRecord.GetTime())
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	tLogRecord, ok := rec.(*apiguard.InputRecord)
	if !ok {
		return nil, storage.ErrFailedTypeAssertion
	}

	if !strings.HasSuffix(tLogRecord.Service, "mquery") {
		log.Warn().Msg("Found non-mquery service record, skipping")
		return []storage.InputRecord{}, nil
	}

	if !tLogRecord.IsCached {
		return []storage.InputRecord{}, nil
	}

	for _, v := range []string{"login", "preflight", "merge-freqs", "speeches", "time-dist-word"} {
		if strings.HasSuffix(tLogRecord.RequestPath, v) {
			log.Debug().Msgf("Skipping virtual apiguard mquery action: %s", v)
			return []storage.InputRecord{}, nil
		}
	}

	return []storage.InputRecord{rec}, nil
}
