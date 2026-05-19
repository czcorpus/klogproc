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

package apiguardKontext

import (
	"klogproc/servicelog/apiguard"
	"strconv"
	"strings"

	"github.com/czcorpus/klogproc-core/storage"
	kontext015 "github.com/czcorpus/klogproc-core/storage/kontext015"
	"github.com/rs/zerolog/log"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeAPIGuardKontext
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*apiguard.InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}

	alignedCorpora, err := tLogRecord.Args.GetStringSlice("align")
	if err != nil {
		return nil, err
	}

	action := tLogRecord.RequestPath
	var userID string
	if tLogRecord.UserID != nil {
		userID = strconv.Itoa(*tLogRecord.UserID)
	}

	r := &kontext015.OutputRecord{
		Type:           storage.AppTypeKontext,
		Action:         action,
		Corpus:         tLogRecord.Args.Get("corpname"),
		AlignedCorpora: alignedCorpora,
		IsCached:       tLogRecord.IsCached,
		IPAddress:      tLogRecord.IPAddress,
		IsAnonymous:    tLogRecord.UserID == nil || storage.UserBelongsToList(*tLogRecord.UserID, t.AnonymousUsers),
		IsAPI:          true, // via APIGuard, only API KonText calls are performed
		IsQuery:        kontext015.IsEntryQuery(action),
		ProcTime:       tLogRecord.ProcTime,
		QueryType:      "", // we cannot decide from URL Query
		UserAgent:      tLogRecord.GetUserAgent(),
		UserID:         userID,
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

	if !strings.HasSuffix(tLogRecord.Service, "kontext") {
		log.Warn().Msg("Found non-kontext service record, skipping")
		return []storage.InputRecord{}, nil
	}

	if !tLogRecord.IsCached {
		return []storage.InputRecord{}, nil
	}

	return []storage.InputRecord{rec}, nil
}
