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
	"time"
)

const poolSize = 1000

// PrevReqPool is a cyclic list containing recently processed
// requests. It is used for searching very similar requests as
// in case of the app 'Mapka', the initialization actually triggers
// two requests we want to count as a single one.
type PrevReqPool struct {
	requests       [poolSize]*OutputRecord
	lastIdx        int
	maxTimeDistSec int
}

// AddItem adds a new record to the pool
func (prp *PrevReqPool) AddItem(rec *OutputRecord) {
	prp.lastIdx = (prp.lastIdx + 1) % poolSize
	prp.requests[prp.lastIdx] = rec
}

func (prp *PrevReqPool) getFirstIdx() int {
	if prp.lastIdx == -1 || prp.requests[(prp.lastIdx+1)%poolSize] == nil {
		return 0
	}
	return (prp.lastIdx + 1) % poolSize
}

// ContainsSimilar tests whethere there is a similar request already
// present. The similarity is tested using:
// 1) IP address
// 2) user agent
// 3) server action
func (prp *PrevReqPool) ContainsSimilar(rec *OutputRecord) bool {
	for i := prp.getFirstIdx(); i%poolSize <= prp.lastIdx; i = (i + 1) % poolSize {
		item := prp.requests[i]
		if item.UserAgent == rec.UserAgent && item.IPAddress == rec.IPAddress &&
			item.Action == rec.Action &&
			rec.GetTime().Sub(item.GetTime()) <= time.Duration(prp.maxTimeDistSec)*time.Second {
			return true
		}
	}
	return false
}

// NewPrevReqPool is a recommended factory for PrevReqPool
func NewPrevReqPool(maxTimeDistSec int) *PrevReqPool {
	return &PrevReqPool{
		lastIdx:        -1,
		maxTimeDistSec: maxTimeDistSec,
	}
}
