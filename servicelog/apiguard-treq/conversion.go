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

package apiguardTreq

import (
	"klogproc/servicelog"
	"klogproc/servicelog/apiguard"
	"klogproc/servicelog/treq"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeAPIGuardTreq
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*apiguard.InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}

	var userID string
	if tLogRecord.UserID != nil {
		userID = strconv.Itoa(*tLogRecord.UserID)
	}

	r := &treq.OutputRecord{
		Type:        servicelog.AppTypeTreq,
		IsAPI:       false, // TODO
		QLang:       tLogRecord.Args.Get("from"),
		SecondLang:  tLogRecord.Args.Get("to"),
		IPAddress:   tLogRecord.IPAddress,
		UserID:      userID,
		IsAnonymous: tLogRecord.UserID == nil,
		Corpus:      "", // TODO
		Subcorpus:   "", // TODO
		IsQuery:     tLogRecord.Args.Get("query") != "",
		IsRegexp:    tLogRecord.Args.Get("regex") == "true",
		IsCaseInsen: tLogRecord.Args.Get("ci") == "true",
		IsMultiWord: tLogRecord.Args.Get("multiword") == "true",
		IsLemma:     tLogRecord.Args.Get("lemma") == "true",
		// GeoIP:       nil, // TODO
	}
	r.SetTime(logRecord.GetTime())
	r.ID = r.GenerateDeterministicID()
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	tLogRecord, ok := rec.(*apiguard.InputRecord)
	if !ok {
		return nil, servicelog.ErrFailedTypeAssertion
	}

	if !strings.HasSuffix(tLogRecord.Service, "treq") {
		log.Debug().Msg("Skipping non-treq service")
		return []servicelog.InputRecord{}, nil
	}

	if !tLogRecord.IsCached {
		log.Debug().Msg("Skipping non-cached treq request")
		return []servicelog.InputRecord{}, nil
	}

	return []servicelog.InputRecord{rec}, nil
}
