// Copyright 2021 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2021 Institute of the Czech National Corpus,
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

package accesslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	entry1 = `10.0.3.50 - janedoe [17/May/2021:06:36:36 +0200] "GET /slovo-v-kostce/search/cs/za%C5%A1kolit?pos=V&lemma=za%C5%A1kolit HTTP/2.0" 200 9218 "https://prirucka.ujc.cas.cz/" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36" rt=0.465`
	entry2 = `10.1.1.15 - - [17/May/2021:08:00:17 +0200] "GET /slovo-v-kostce/assets/alt-view.svg HTTP/2.0" 200 1793 "https://www.korpus.cz/slovo-v-kostce/search/cs/v%C3%BDbuch" "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"`
)

func TestRandomEntry(t *testing.T) {
	parser := LineParser{}
	tokens, _ := parser.tokenize(entry1)
	assert.Equal(t, "10.0.3.50", tokens[0])
	assert.Equal(t, "-", tokens[1])
	assert.Equal(t, "janedoe", tokens[2])
	assert.Equal(t, "17/May/2021:06:36:36 +0200", tokens[3])
	assert.Equal(t, "GET /slovo-v-kostce/search/cs/za%C5%A1kolit?pos=V&lemma=za%C5%A1kolit HTTP/2.0", tokens[4])
	assert.Equal(t, "200", tokens[5])
	assert.Equal(t, "9218", tokens[6])
	assert.Equal(t, "https://prirucka.ujc.cas.cz/", tokens[7])
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36", tokens[8])
	assert.Equal(t, "rt=0.465", tokens[9])
}

// TestRandomEntryWithoutRt tests parsing of an entry without processing time information.
func TestRandomEntryWithoutRt(t *testing.T) {
	parser := LineParser{}
	tokens, _ := parser.tokenize(entry2)
	assert.Equal(t, 10, len(tokens))
	assert.Equal(t, "", tokens[len(tokens)-1])
}
