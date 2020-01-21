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
	"github.com/czcorpus/klogproc/conversion/kwords"
	"github.com/czcorpus/klogproc/conversion/morfio"
	"github.com/czcorpus/klogproc/conversion/shiny"
	"github.com/czcorpus/klogproc/conversion/ske"
	"github.com/czcorpus/klogproc/conversion/syd"
	"github.com/czcorpus/klogproc/conversion/treq"
	"github.com/czcorpus/klogproc/conversion/wag"
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

type morfioLineParser struct {
	lp *morfio.LineParser
}

func (parser *morfioLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type kwordsLineParser struct {
	lp *kwords.LineParser
}

func (parser *kwordsLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type skeLineParser struct {
	lp *ske.LineParser
}

func (parser *skeLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type wagLineParser struct {
	lp *wag.LineParser
}

func (parser *wagLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

type shinyLineParser struct {
	lp *shiny.LineParser
}

func (parser *shinyLineParser) ParseLine(s string, lineNum int, localTimezone string) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

// ------------------------------------

// NewLineParser creates a parser for individual lines of a respective appType
func NewLineParser(appType string, appErrRegister conversion.AppErrorRegister) (LineParser, error) {
	switch appType {
	case conversion.AppTypeKontext:
		return &kontextLineParser{lp: kontext.NewLineParser(appErrRegister)}, nil
	case conversion.AppTypeSyd:
		return &sydLineParser{lp: &syd.LineParser{}}, nil
	case conversion.AppTypeTreq:
		return &treqLineParser{lp: &treq.LineParser{}}, nil
	case conversion.AppTypeMorfio:
		return &morfioLineParser{lp: &morfio.LineParser{}}, nil
	case conversion.AppTypeKwords:
		return &kwordsLineParser{lp: &kwords.LineParser{}}, nil
	case conversion.AppTypeSke:
		return &skeLineParser{lp: &ske.LineParser{}}, nil
	case conversion.AppTypeWag:
		return &wagLineParser{lp: &wag.LineParser{}}, nil
	case conversion.AppTypeCalc, conversion.AppTypeLists:
		return &shinyLineParser{lp: &shiny.LineParser{}}, nil
	default:
		return nil, fmt.Errorf("Parser not found for application type %s", appType)
	}
}
