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

	"klogproc/conversion"
	"klogproc/conversion/apiguard"
	"klogproc/conversion/kontext013"
	"klogproc/conversion/kontext015"
	"klogproc/conversion/kontext018"
	"klogproc/conversion/korpusdb"
	"klogproc/conversion/kwords"
	"klogproc/conversion/mapka"
	"klogproc/conversion/mapka2"
	"klogproc/conversion/mapka3"
	"klogproc/conversion/masm"
	"klogproc/conversion/morfio"
	"klogproc/conversion/shiny"
	"klogproc/conversion/ske"
	"klogproc/conversion/syd"
	"klogproc/conversion/treq"
	"klogproc/conversion/wag07"

	"klogproc/conversion/wag06"
	"klogproc/conversion/wsserver"
)

// ------------------------------------

type apiguardLineParser struct {
	lp *apiguard.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *apiguardLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

// kontext013LineParser wraps kontext-specific parser into a general form as required
// by core of the klogproc
type kontext013LineParser struct {
	lp *kontext013.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *kontext013LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

// kontext015LineParser wraps kontext-specific parser into a general form as required
// by core of the klogproc
type kontext015LineParser struct {
	lp *kontext015.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *kontext015LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

// kontext018LineParser wraps kontext-specific parser into a general form as required
// by core of the klogproc
type kontext018LineParser struct {
	lp *kontext018.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *kontext018LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type kwordsLineParser struct {
	lp *kwords.LineParser
}

func (parser *kwordsLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type korpusDBLineParser struct {
	lp *korpusdb.LineParser
}

func (parser *korpusDBLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type mapkaLineParser struct {
	lp *mapka.LineParser
}

func (parser *mapkaLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type mapka2LineParser struct {
	lp *mapka2.LineParser
}

func (parser *mapka2LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type mapka3LineParser struct {
	lp *mapka3.LineParser
}

func (parser *mapka3LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type morfioLineParser struct {
	lp *morfio.LineParser
}

func (parser *morfioLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type shinyLineParser struct {
	lp *shiny.LineParser
}

func (parser *shinyLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type skeLineParser struct {
	lp *ske.LineParser
}

func (parser *skeLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type sydLineParser struct {
	lp *syd.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *sydLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type treqLineParser struct {
	lp *treq.LineParser
}

func (parser *treqLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type wag06LineParser struct {
	lp *wag06.LineParser
}

func (parser *wag06LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type wag07LineParser struct {
	lp *wag07.LineParser
}

func (parser *wag07LineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type wsserverLineParser struct {
	lp *wsserver.LineParser
}

func (parser *wsserverLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

type masmLineParser struct {
	lp *masm.LineParser
}

func (parser *masmLineParser) ParseLine(s string, lineNum int64) (conversion.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum)
}

// ------------------------------------

// NewLineParser creates a parser for individual lines of a respective appType
func NewLineParser(appType string, version string, appErrRegister conversion.AppErrorRegister) (LineParser, error) {
	switch appType {
	case conversion.AppTypeAPIGuard:
		return &apiguardLineParser{lp: &apiguard.LineParser{}}, nil
	case conversion.AppTypeAkalex, conversion.AppTypeCalc, conversion.AppTypeLists,
		conversion.AppTypeQuitaUp, conversion.AppTypeGramatikat:
		return &shinyLineParser{lp: &shiny.LineParser{}}, nil
	case conversion.AppTypeKontext, conversion.AppTypeKontextAPI:
		switch version {
		case "0.13", "0.14":
			return &kontext013LineParser{lp: kontext013.NewLineParser(appErrRegister)}, nil
		case "0.15", "0.16", "0.17":
			return &kontext015LineParser{lp: kontext015.NewLineParser(appErrRegister)}, nil
		case "0.18":
			return &kontext018LineParser{lp: kontext018.NewLineParser()}, nil
		default:
			return nil, fmt.Errorf("cannot find parser - unsupported version of KonText specified: %s", version)
		}
	case conversion.AppTypeKwords:
		return &kwordsLineParser{lp: &kwords.LineParser{}}, nil
	case conversion.AppTypeKorpusDB:
		return &korpusDBLineParser{lp: &korpusdb.LineParser{}}, nil
	case conversion.AppTypeMapka:
		switch version {
		case "1":
			return &mapkaLineParser{lp: &mapka.LineParser{}}, nil
		case "2":
			return &mapka2LineParser{lp: &mapka2.LineParser{}}, nil
		case "3":
			return &mapka3LineParser{lp: &mapka3.LineParser{}}, nil
		default:
			return nil, fmt.Errorf("cannot find parser - unsupported version of Mapka specified: %s", version)
		}
	case conversion.AppTypeMorfio:
		return &morfioLineParser{lp: &morfio.LineParser{}}, nil
	case conversion.AppTypeSke:
		return &skeLineParser{lp: &ske.LineParser{}}, nil
	case conversion.AppTypeSyd:
		return &sydLineParser{lp: &syd.LineParser{}}, nil
	case conversion.AppTypeTreq:
		return &treqLineParser{lp: &treq.LineParser{}}, nil
	case conversion.AppTypeWag:
		switch version {
		case "0.6":
			return &wag06LineParser{lp: &wag06.LineParser{}}, nil
		case "0.7":
			return &wag07LineParser{lp: &wag07.LineParser{}}, nil
		default:
			return nil, fmt.Errorf("cannot find parser - unsupported version of WaG specified: %s", version)
		}
	case conversion.AppTypeWsserver:
		return &wsserverLineParser{lp: &wsserver.LineParser{}}, nil
	case conversion.AppTypeMasm:
		return &masmLineParser{lp: &masm.LineParser{}}, nil
	default:
		return nil, fmt.Errorf("Parser not found for application type %s", appType)
	}
}
