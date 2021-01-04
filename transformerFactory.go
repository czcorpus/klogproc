// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

package main

import (
	"fmt"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/conversion/kontext013"
	"github.com/czcorpus/klogproc/conversion/kontext015"
	"github.com/czcorpus/klogproc/conversion/korpusdb"
	"github.com/czcorpus/klogproc/conversion/kwords"
	"github.com/czcorpus/klogproc/conversion/mapka"
	"github.com/czcorpus/klogproc/conversion/morfio"
	"github.com/czcorpus/klogproc/conversion/shiny"
	"github.com/czcorpus/klogproc/conversion/ske"
	"github.com/czcorpus/klogproc/conversion/syd"
	"github.com/czcorpus/klogproc/conversion/treq"
	"github.com/czcorpus/klogproc/conversion/wag"
	"github.com/czcorpus/klogproc/conversion/wsserver"
	"github.com/czcorpus/klogproc/users"
)

// ------------------------------------

// konText013Transformer is a simple type-safe wrapper for kontext v 0.13.x app type log transformer
type konText013Transformer struct {
	t *kontext013.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *konText013Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kontext013.InputRecord)
	if ok {
		return k.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

// konText015Transformer is a simple type-safe wrapper for kontext 0.15.x app type log transformer
type konText015Transformer struct {
	t *kontext015.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *konText015Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kontext015.InputRecord)
	if ok {
		return k.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

type kwordsTransformer struct {
	t *kwords.Transformer
}

// Transform transforms KWords app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *kwordsTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kwords.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KWords transformer %T", logRec)
}

// ------------------------------------

type korpusDBTransformer struct {
	t *korpusdb.Transformer
}

// Transform transforms KorpusDB app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *korpusDBTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*korpusdb.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

type mapkaTransformer struct {
	t *mapka.Transformer
}

// Transform transforms Mapka app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *mapkaTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*mapka.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Mapka transformer %T", logRec)
}

// ------------------------------------

type morfioTransformer struct {
	t *morfio.Transformer
}

// Transform transforms Morfio app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *morfioTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*morfio.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Morfio transformer %T", logRec)
}

// ------------------------------------

type skeTransformer struct {
	t *ske.Transformer
}

// Transform transforms SkE app log record (= web access log) types as general InputRecord
// In case of type mismatch, error is returned.
func (s *skeTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*ske.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by SkE transformer %T", logRec)
}

// ------------------------------------

type shinyTransformer struct {
	t *shiny.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *shinyTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*shiny.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Shiny transformer %T", logRec)
}

// ------------------------------------

type sydTransformer struct {
	t *syd.Transformer
}

// Transform transforms SyD app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *sydTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*syd.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by SyD transformer %T", logRec)
}

// ------------------------------------

type treqTransformer struct {
	t *treq.Transformer
}

// Transform transforms Treq app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *treqTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*treq.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Treq transformer %T", logRec)
}

// ------------------------------------

type wagTransformer struct {
	t *wag.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wagTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*wag.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by WaG transformer %T", logRec)
}

// ------------------------------------

type wsserverTransformer struct {
	t *wsserver.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wsserverTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*wsserver.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by WSServer transformer %T", logRec)
}

// ------------------------------------

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(appType string, version string, userMap *users.UserMap) (conversion.LogItemTransformer, error) {
	switch appType {
	case conversion.AppTypeCalc, conversion.AppTypeLists, conversion.AppTypeQuitaUp:
		return &shinyTransformer{t: shiny.NewTransformer()}, nil
	case conversion.AppTypeKontext, conversion.AppTypeKontextAPI:
		switch version {
		case "0.13":
			return &konText013Transformer{t: &kontext013.Transformer{}}, nil
		case "0.15":
			return &konText015Transformer{t: &kontext015.Transformer{}}, nil
		default:
			return nil, fmt.Errorf("Cannot create transformer, unsupported KonText version: %s", version)
		}
	case conversion.AppTypeKwords:
		return &kwordsTransformer{t: &kwords.Transformer{}}, nil
	case conversion.AppTypeKorpusDB:
		return &korpusDBTransformer{t: &korpusdb.Transformer{}}, nil
	case conversion.AppTypeMapka:
		return &mapkaTransformer{t: mapka.NewTransformer()}, nil
	case conversion.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{}}, nil
	case conversion.AppTypeSke:
		return &skeTransformer{t: ske.NewTransformer(userMap)}, nil
	case conversion.AppTypeSyd:
		return &sydTransformer{t: syd.NewTransformer(version)}, nil
	case conversion.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{}}, nil
	case conversion.AppTypeWag:
		return &wagTransformer{t: &wag.Transformer{}}, nil
	case conversion.AppTypeWsserver:
		return &wsserverTransformer{t: &wsserver.Transformer{}}, nil
	default:
		return nil, fmt.Errorf("Cannot find log transformer for app type %s", appType)
	}
}
