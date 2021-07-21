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
	"fmt"
	"strings"

	"github.com/czcorpus/klogproc/conversion"
)

// LineParser is a parser for reading Treq application log
// which is basically a TAB separated list of items.
type LineParser struct {
}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
// Format (adopted from Treq production version 2019-07-15)
// type D:  leftLang[TAB]rightLang[TAB]viceslovne[TAB]lemma[TAB]dataPack[TAB]regularni[TAB]caseInsen[TAB]hledejCo[TAB]...[TAB]
// type L:  Gleft   [TAB]Gright   [TAB               ]lemma[TAB]dataPack[TAB].........[TAB].........[TAB]Gquery1[TAB]Gquery2
// please note that for the two query types, the columns are shifted
func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {

	items := strings.Split(s, "\t")
	var err error

	if len(items) >= 10 && items[3] == "L" {
		return &InputRecord{
			Datetime:   items[0],
			IPAddress:  items[1],
			UserID:     items[2],
			QType:      items[3],
			QLang:      items[4],
			SecondLang: items[5],
			// No multiWord info in case of 'L' query; not even an empty col
			IsLemma:   items[6],
			Subcorpus: strings.ToUpper(items[7]), // we have to normalize because of Treq
			// No IsRegexp; not even an empty col
			// No IsCaseInsen; not even an empty col
			Query:  items[8],
			Query2: items[9],
		}, err

	} else if len(items) >= 12 && items[3] == "D" {
		return &InputRecord{
			Datetime:    items[0],
			IPAddress:   items[1],
			UserID:      items[2],
			QType:       items[3],
			QLang:       items[4],
			SecondLang:  items[5],
			IsMultiWord: items[6],
			IsLemma:     items[7],
			Subcorpus:   strings.ToUpper(items[8]), // we have to normalize because of Treq
			IsRegexp:    items[9],
			IsCaseInsen: items[10],
			Query:       items[11],
			// No query2 in case of the 'D' query
		}, err
	}
	return nil, conversion.NewLineParsingError(
		lineNum, fmt.Sprintf("Invalid line format. Expecting min. 10 (type L) or min. 12 (type D) tab-separated items, found %d", len(items)))
}
