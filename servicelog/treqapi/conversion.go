// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
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

package treqapi

import (
	"fmt"
	"strconv"

	"github.com/czcorpus/klogproc-core/scripting"
	"github.com/czcorpus/klogproc-core/storage"
	treqCore "github.com/czcorpus/klogproc-core/storage/treq"
	lua "github.com/yuin/gopher-lua"
)

// Transformer converts a Treq log record to a destination format
type Transformer struct {
	AnonymousUsers []int
}

func (t *Transformer) AppType() string {
	return storage.AppTypeTreq
}

func (t *Transformer) Transform(
	logRecord storage.InputRecord,
) (storage.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(storage.ErrFailedTypeAssertion)
	}

	userID := -1
	if tLogRecord.UserID != "-" {
		if uid, err := strconv.Atoi(tLogRecord.UserID); err == nil {
			userID = uid

		} else {
			return nil, fmt.Errorf(
				"failed to convert user ID '%s': %w", tLogRecord.UserID, err)
		}
	}

	out := &treqCore.OutputRecord{
		Type:        "treq",
		IsAPI:       true,
		IsQuery:     true,
		QLang:       tLogRecord.From,
		SecondLang:  tLogRecord.To,
		IPAddress:   tLogRecord.IP,
		UserID:      tLogRecord.UserID,
		IsAnonymous: userID == -1 || storage.UserBelongsToList(userID, t.AnonymousUsers),
		IsRegexp:    tLogRecord.Regex,
		IsCaseInsen: tLogRecord.CI,
		IsMultiWord: tLogRecord.Multiword,
		IsLemma:     tLogRecord.Lemma,
	}
	out.SetTime(tLogRecord.GetTime())
	// !!! Due to unique hash generation and a bug in older records with isQuery:false, we have
	// We have to ensure that IDs are generated consistently, so we keep isQuery:false
	// when calculating the hash.
	out_compat := *out
	out_compat.IsQuery = false
	out.ID = out_compat.GenerateDeterministicID()
	return out, nil
}

func (t *Transformer) SetOutputProperty(rec storage.OutputRecord, name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec storage.InputRecord, prevRecs storage.ServiceLogBuffer,
) ([]storage.InputRecord, error) {
	return []storage.InputRecord{rec}, nil
}
