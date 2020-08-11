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
	"strconv"
	"time"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/users"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
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
		IsQuery:     false, // TODO
		UserID:      strconv.Itoa(userID),
		Action:      logRecord.Action,
		Path:        logRecord.Path,
		ProcTime:    logRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(userMap *users.UserMap) *Transformer {
	return &Transformer{}
}
