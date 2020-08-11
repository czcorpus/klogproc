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

package kwords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	line := `2019-07-08T18:16:23+02:00	192.168.1.65	99	7	T	2358	X	4869	2	3	4	5	6`
	// please note that the last 5 values tested are semantically incorrect as they should be 0/1 encoded booleans
	// but for the sake of this test we want to distinguish between them which means we need more values.
	p := LineParser{}
	rec, err := p.ParseLine(line, 71)
	assert.Nil(t, err)
	// kwords ignore timezone correction
	assert.Equal(t, "2019-07-08T18:16:23+02:00", rec.Datetime)
	assert.Equal(t, "192.168.1.65", rec.IPAddress)
	assert.Equal(t, "99", rec.UserID)
	assert.Equal(t, "7", rec.NumFiles)
	assert.Equal(t, "T", rec.TargetInputType)
	assert.Equal(t, "2358", rec.TargetLength)
	assert.Equal(t, "X", rec.Corpus)
	assert.Equal(t, "4869", rec.RefLength)
	assert.Equal(t, "2", rec.Pronouns)
	assert.Equal(t, "3", rec.Prep)
	assert.Equal(t, "4", rec.Con)
	assert.Equal(t, "5", rec.Num)
	assert.Equal(t, "6", rec.CaseInsensitive)

}
