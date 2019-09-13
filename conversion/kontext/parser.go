// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package kontext

import (
	"fmt"
	"regexp"

	"github.com/czcorpus/klogproc/conversion"
)

var (
	logLineRegexp  = regexp.MustCompile("^.+\\sINFO:\\s+(\\{.+)$")
	lineTypeRegexp = regexp.MustCompile("^.+\\s([A-Z]+):\\s+.+$")
)

func parseRawLine(s string) string {
	srch := logLineRegexp.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func getLineType(s string) string {
	srch := lineTypeRegexp.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

// LineParser is a parser for reading KonText application logs
type LineParser struct{}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	jsonLine := parseRawLine(s)
	if jsonLine != "" {
		return ImportJSONLog([]byte(jsonLine), localTimezone)

	} else if tp := getLineType(s); tp == "QUERY" {
		return nil, fmt.Errorf("Failed to process QUERY entry: %s", s)

	} else {
		return nil, conversion.NewLineParsingError(lineNum, fmt.Sprintf("ignored non-query entry"))
	}
}
