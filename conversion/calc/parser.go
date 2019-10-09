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

package calc

import (
	"encoding/json"
)

// LineParser is a parser for reading KonText application logs
type LineParser struct {
}

func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	rec := &InputRecord{}
	err := json.Unmarshal([]byte(s), rec)
	if err != nil {
		return rec, err
	}
	if rec.TS[len(rec.TS)-1] == 'Z' {
		rec.TS = rec.TS[:len(rec.TS)-1] + "+01:00"
	}
	return rec, nil
}
