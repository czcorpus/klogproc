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

package mapka

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

// createID creates an idempotent ID of rec based on its properties.
func createID(rec *OutputRecord) string {
	str := rec.Type + rec.Path + rec.Datetime + rec.IPAddress + rec.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	prevReqs   *PrevReqPool
	numSimilar int
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*OutputRecord, error) {
	userID := -1

	r := &OutputRecord{
		Type:        recType,
		time:        logRecord.GetTime(),
		Datetime:    logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:   logRecord.Request.RemoteAddr,
		UserAgent:   logRecord.Request.HTTPUserAgent,
		IsAnonymous: userID == -1 || conversion.UserBelongsToList(userID, anonymousUsers),
		IsQuery:     false,
		UserID:      strconv.Itoa(userID),
		Action:      logRecord.Action,
		Path:        logRecord.Path,
		ProcTime:    logRecord.ProcTime,
		Params:      logRecord.Params,
	}
	r.ID = createID(r)
	if t.prevReqs.ContainsSimilar(r) && r.Action == "overlay" || r.Action == "text" {
		r.IsQuery = true
	}
	t.prevReqs.AddItem(r)
	return r, nil
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer() *Transformer {
	return &Transformer{
		prevReqs: NewPrevReqPool(5),
	}
}
