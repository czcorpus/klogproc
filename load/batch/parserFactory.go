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

package batch

import (
	"fmt"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/conversion/kontext"
	"github.com/czcorpus/klogproc/conversion/syd"
	"github.com/czcorpus/klogproc/conversion/treq"
)

// ------------------------------------

// kontextLineParser wraps kontext-specific parser into a general form as required
// by core of the klogproc
type kontextLineParser struct {
	lp *kontext.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *kontextLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type sydLineParser struct {
	lp *syd.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *sydLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type treqLineParser struct {
	lp *treq.LineParser
}

func (parser *treqLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

// NewLineParser creates a parser for individual lines of a respective appType
func NewLineParser(appType string) (LineParser, error) {
	switch appType {
	case conversion.AppTypeKontext:
		return &kontextLineParser{lp: &kontext.LineParser{}}, nil
	case conversion.AppTypeSyd:
		return &sydLineParser{lp: &syd.LineParser{}}, nil
	case conversion.AppTypeTreq:
		return &treqLineParser{lp: &treq.LineParser{}}, nil
	default:
		return nil, fmt.Errorf("Parser not found for application type %s", appType)
	}
}