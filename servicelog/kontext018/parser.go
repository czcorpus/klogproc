// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package kontext018

import (
	"encoding/json"
	"klogproc/servicelog"
)

// LineParser is a parser for reading KonText application logs
type LineParser struct {
}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
func (lp *LineParser) ParseLine(s string, lineNum int64) (*QueryInputRecord, error) {
	var record QueryInputRecord
	err := json.Unmarshal([]byte(s), &record)
	if err != nil {
		return nil, servicelog.NewStreamedLineParsingError(s, "json Unmarshal error")
	}
	if record.Logger == "QUERY" {
		record.isProcessable = true
	}
	return &record, nil
}

// NewLineParser is a factory for LineParser
func NewLineParser() *LineParser {
	return &LineParser{}
}
