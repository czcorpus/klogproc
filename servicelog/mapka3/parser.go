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

package mapka3

import (
	"encoding/json"
)

// LineParser is a parser for reading Mapka application logs
type LineParser struct {
}

// ParseLine parses a HTTP access log format line
func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {
	parsed := new(InputRecord)
	err := json.Unmarshal([]byte(s), parsed)
	if err != nil {
		return &InputRecord{isProcessable: false}, err
	}
	parsed.isProcessable = true
	return parsed, nil
}
