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

package mapka3

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"klogproc/analysis/clustering"
	"klogproc/load"
	"klogproc/scripting"
	"klogproc/servicelog"

	lua "github.com/yuin/gopher-lua"
)

// createID creates an idempotent ID of rec based on its properties.
func createID(rec *OutputRecord, tzShiftMin int) string {
	str := rec.Type +
		rec.Path +
		rec.GetTime().Add(time.Minute*time.Duration(tzShiftMin)).Format(time.RFC3339Nano) +
		rec.IPAddress +
		rec.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	bufferConf     *load.BufferConf
	analyzer       servicelog.Preprocessor
	excludeIPList  servicelog.ExcludeIPList
	anonymousUsers []int
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeMapka
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	r := &OutputRecord{
		Type:      t.AppType(),
		time:      tLogRecord.GetTime(),
		Datetime:  tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress: tLogRecord.GetClientIP().String(),
		UserAgent: tLogRecord.GetUserAgent(),
		IsAnonymous: tLogRecord.Extra.UserID == "" ||
			servicelog.UserBelongsToList(tLogRecord.Extra.UserID, t.anonymousUsers),
		Action:      "interaction",
		UserID:      tLogRecord.Extra.UserID,
		ClusterSize: tLogRecord.clusterSize,
	}
	if r.ClusterSize > 0 {
		r.IsQuery = true
	}
	r.ID = createID(r, tzShiftMin)
	return r, nil
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	return scripting.ErrScriptingNotSupported
}

func (t *Transformer) HistoryLookupItems() int {
	return t.bufferConf.HistoryLookupItems
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord,
	prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if t.excludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return t.analyzer.Preprocess(rec, prevRecs)
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(
	bufferConf *load.BufferConf,
	excludeIPList servicelog.ExcludeIPList,
	anonymousUsers []int,
	realtimeClock bool,
) *Transformer {
	return &Transformer{
		bufferConf:     bufferConf,
		excludeIPList:  excludeIPList,
		anonymousUsers: anonymousUsers,
		analyzer:       clustering.NewAnalyzer[*InputRecord]("mapka", bufferConf, realtimeClock),
	}
}
