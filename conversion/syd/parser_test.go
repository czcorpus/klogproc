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

func TestParseLine(t *testing.T) {
	line := `2019-07-26T02:43:16+02:00	2a00:1028:8386:6532:5d6c:d8cd:ceab:40da	-	Q	N	VJLAVCnL	S	R`
	p := LineParser{}
	rec, err := p.ParseLine(line, 71, "+01:00")
	assert.Nil(t, err)
	assert.Equal(t, "2019-07-26T02:43:16+02:00", rec.Datetime)
	assert.Equal(t, "2a00:1028:8386:6532:5d6c:d8cd:ceab:40da", rec.IPAddress)
	assert.Equal(t, "-", rec.UserID)
	assert.Equal(t, "Q", rec.KeyReq)
	assert.Equal(t, "N", rec.KeyUsed)
	assert.Equal(t, "VJLAVCnL", rec.Key)
	assert.Equal(t, "S", rec.Ltool)
	assert.Equal(t, "R", rec.RunScript)
}
