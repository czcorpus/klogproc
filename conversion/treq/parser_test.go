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

package treq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// note: the parser has been tested with production Treq which does not
// match the Git repository version. When debugging the parser, please
// check index.php first. The md5 checksum of the one used for this
// parser is ba913394208ade35facff8076917da16.

func TestParseLineTypeD(t *testing.T) {
	line := `2019-07-24T11:52:42+02:00	127.0.0.1	1531	D	cs	en	1	2	ACQUIS|EUROPARL|CORE	3	4	mocnost`
	/*
		respective Treq code:
		$query = $_SESSION["leftLang"] . "\t" . $_SESSION["rightLang"] . "\t" . $_SESSION["viceslovne"] . "\t" .
		$_SESSION["lemma"] . "\t" . $DdataPack . "\t" . $_SESSION["regularni"] . "\t" . $_SESSION["caseInsen"] .
		"\t" . $_SESSION["hledejCo"] . "\t\t";
	*/
	p := LineParser{}
	rec, err := p.ParseLine(line, 71, "+01:00")
	assert.Nil(t, err)
	assert.Equal(t, "2019-07-24T11:52:42+02:00", rec.Datetime)
	assert.Equal(t, "127.0.0.1", rec.IPAddress)
	assert.Equal(t, "1531", rec.UserID)
	assert.Equal(t, "D", rec.QType)
	assert.Equal(t, "cs", rec.QLang)
	assert.Equal(t, "en", rec.SecondLang)
	assert.Equal(t, "1", rec.IsMultiWord)
	assert.Equal(t, "2", rec.IsLemma) // 2 is semantically incorrect but we need an unique value
	assert.Equal(t, "", rec.Corpus)
	assert.Equal(t, "ACQUIS|EUROPARL|CORE", rec.Subcorpus)
	assert.Equal(t, "3", rec.IsRegexp)    // 3 is semantically incorrect but we need an unique value
	assert.Equal(t, "4", rec.IsCaseInsen) // 4 is semantically incorrect but we need an unique value
	assert.Equal(t, "mocnost", rec.Query)
	assert.Equal(t, "", rec.Query2)
}

func TestParseLineTypeL(t *testing.T) {
	line := `2017-03-26T14:27:27+02:00	127.0.0.1	-	L	en	cs	1	PressEurop|Syndicate|Subtitles	gear	rychlost`
	/*
			respective Treq logging code:
		    $query = $Gleft . "\t" . $Gright . "\t" . $Glemma . "\t" . $GdataPack . "\t\t\t" . $Gquery1 . "\t" . $Gquery2;
	*/
	p := LineParser{}
	rec, err := p.ParseLine(line, 71, "+01:00")
	assert.Nil(t, err)
	assert.Equal(t, "2017-03-26T14:27:27+02:00", rec.Datetime)
	assert.Equal(t, "127.0.0.1", rec.IPAddress)
	assert.Equal(t, "-", rec.UserID)
	assert.Equal(t, "L", rec.QType)
	assert.Equal(t, "en", rec.QLang)
	assert.Equal(t, "cs", rec.SecondLang)
	assert.Equal(t, "", rec.IsMultiWord)
	assert.Equal(t, "1", rec.IsLemma)
	assert.Equal(t, "", rec.Corpus)
	assert.Equal(t, "PRESSEUROP|SYNDICATE|SUBTITLES", rec.Subcorpus)
	assert.Equal(t, "", rec.IsRegexp)
	assert.Equal(t, "", rec.IsCaseInsen)
	assert.Equal(t, "gear", rec.Query)
	assert.Equal(t, "rychlost", rec.Query2)

}
