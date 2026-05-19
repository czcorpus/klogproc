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

package apiguardKwords

import (
	"klogproc/servicelog/apiguard"
	"strconv"
	"strings"

	"github.com/czcorpus/klogproc-core/storage"
	kwords2Core "github.com/czcorpus/klogproc-core/storage/kwords2"
	"github.com/rs/zerolog/log"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeAPIGuardTreq
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*apiguard.InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}

	var userID *string
	if tLogRecord.UserID != nil {
		tmp := strconv.Itoa(*tLogRecord.UserID)
		userID = &tmp
	}

	// note that in case of APIGuard logs for Kwords, we cannot
	// fetch most of the attributes as APIGuard does not know details
	// about KWords requests and its logging procedure knows only
	// request URLs.
	r := &kwords2Core.OutputRecord{
		Type:        storage.AppTypeKwords,
		IPAddress:   tLogRecord.IPAddress,
		IsCached:    tLogRecord.IsCached,
		UserID:      userID,
		IsAnonymous: tLogRecord.UserID == nil || storage.UserBelongsToList(*tLogRecord.UserID, t.AnonymousUsers),
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

	if !strings.HasSuffix(tLogRecord.Service, "kwords") {
		log.Warn().Msg("Found non-kwords service record, skipping")
		return []storage.InputRecord{}, nil
	}

	if !tLogRecord.IsCached {
		return []storage.InputRecord{}, nil
	}

	return []storage.InputRecord{rec}, nil
}
