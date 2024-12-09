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

package syd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformDia(t *testing.T) {
	tmr := NewTransformer("0.1", []int{0, 1})
	rec := &InputRecord{
		UserID: "30",
		Ltool:  "D",
	}
	outRec, err := tmr.Transform(rec, 0)
	tOutRec, ok := outRec.(*OutputRecord)
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Contains(t, tOutRec.Corpus, "diakon")
	assert.Equal(t, 1, len(tOutRec.Corpus))
}

func TestTransformSync(t *testing.T) {
	tmr := NewTransformer("0.1", []int{0, 1})
	rec := &InputRecord{
		UserID: "30",
		Ltool:  "S",
	}
	outRec, err := tmr.Transform(rec, 0)
	tOutRec, ok := outRec.(*OutputRecord)
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Contains(t, tOutRec.Corpus, "syn2010")
	assert.Contains(t, tOutRec.Corpus, "oral_v2")
	assert.Contains(t, tOutRec.Corpus, "ksk-dopisy")
	assert.Equal(t, 3, len(tOutRec.Corpus))
}

func TestAcceptsDashAsUserID(t *testing.T) {
	tmr := NewTransformer("0.1", []int{0, 1})
	rec := &InputRecord{
		UserID: "-",
	}
	outRec, err := tmr.Transform(rec, 0)
	assert.Nil(t, err)
	tOutRec, ok := outRec.(*OutputRecord)
	assert.True(t, ok)
	assert.Nil(t, tOutRec.UserID)
}

func TestAnonymousUserDetection(t *testing.T) {
	tmr := NewTransformer("0.1", []int{26, 27})

	rec := &InputRecord{
		UserID: "27",
	}
	outRec, err := tmr.Transform(rec, 0)
	assert.Nil(t, err)
	tOutRec, ok := outRec.(*OutputRecord)
	assert.True(t, ok)
	assert.True(t, tOutRec.IsAnonymous)

	rec = &InputRecord{
		UserID: "28",
	}
	outRec, err = tmr.Transform(rec, 0)
	assert.Nil(t, err)
	tOutRec, ok = outRec.(*OutputRecord)
	assert.True(t, ok)
	assert.False(t, tOutRec.IsAnonymous)
}
