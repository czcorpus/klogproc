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

package morfio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLine(t *testing.T) {
	line := `2019-07-25T15:40:54+02:00	127.0.0.1	953	Q	N	brBRepBf	R	syn2015	0	lemma	word	yes	 (.+) ý . 	 (.+) ák . `
	p := LineParser{}
	rec, err := p.ParseLine(line, 71)
	assert.Nil(t, err)
	assert.Equal(t, "2019-07-25T15:40:54+02:00", rec.Datetime)
	assert.Equal(t, "127.0.0.1", rec.IPAddress)
	assert.Equal(t, "953", rec.UserID)
	assert.Equal(t, "Q", rec.KeyReq)
	assert.Equal(t, "N", rec.KeyUsed)
	assert.Equal(t, "brBRepBf", rec.Key)
	assert.Equal(t, "R", rec.RunScript)
	assert.Equal(t, "syn2015", rec.Corpus)
	assert.Equal(t, "0", rec.MinFreq)
	assert.Equal(t, "lemma", rec.InputAttr)
	assert.Equal(t, "word", rec.OutputAttr)
	assert.Equal(t, "yes", rec.CaseInsensitive)

}
