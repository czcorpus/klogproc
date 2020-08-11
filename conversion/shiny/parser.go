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

package shiny

import (
	"encoding/json"
	"regexp"
)

var (
	tzSrch = regexp.MustCompile("\\d[-+][0-1]\\d:[0-5]\\d$")
)

// LineParser is a parser for reading KonText application logs
type LineParser struct {
	AnonymousUserID int
}

func (lp *LineParser) ParseLine(s string, lineNum int) (*InputRecord, error) {
	rec := &InputRecord{}
	err := json.Unmarshal([]byte(s), rec)
	if err != nil {
		return rec, err
	}
	if rec.TS[len(rec.TS)-1] == 'Z' { // UTC time
		rec.TS = rec.TS[:len(rec.TS)-1]

	} else if tzSrch.FindString(rec.TS) != "" {
		rec.TS = rec.TS[:len(rec.TS)-5]
	}
	return rec, nil
}
