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

	"klogproc/conversion"
	"klogproc/conversion/kontext013"
	"klogproc/conversion/kontext015"
	"klogproc/conversion/kontext018"
	"klogproc/conversion/korpusdb"
	"klogproc/conversion/kwords"
	"klogproc/conversion/mapka"
	"klogproc/conversion/mapka2"
	"klogproc/conversion/masm"
	"klogproc/conversion/morfio"
	"klogproc/conversion/shiny"
	"klogproc/conversion/ske"
	"klogproc/conversion/syd"
	"klogproc/conversion/treq"
	"klogproc/conversion/wag06"
	"klogproc/conversion/wag07"
	"klogproc/conversion/wsserver"
	"klogproc/users"
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
	return nil, fmt.Errorf("invalid type for conversion by KonText transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

// konText018Transformer is a simple type-safe wrapper for kontext 0.18.x app type log transformer
type konText018Transformer struct {
	t *kontext018.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *konText018Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kontext018.QueryInputRecord)
	if ok {
		return k.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by KonText transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by KWords transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by KonText transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by Mapka transformer %T", logRec)
}

// ------------------------------------

type mapka2Transformer struct {
	t *mapka2.Transformer
}

// Transform transforms Mapka (v2) app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *mapka2Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*mapka2.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by Mapka2 transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by Morfio transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by SkE transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by Shiny transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by SyD transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by Treq transformer %T", logRec)
}

// ------------------------------------

type wag06Transformer struct {
	t *wag06.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wag06Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*wag06.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by WaG 0.6 transformer %T", logRec)
}

// ------------------------------------

type wag07Transformer struct {
	t *wag07.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wag07Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*wag07.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by WaG 0.7 transformer %T", logRec)
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
	return nil, fmt.Errorf("invalid type for conversion by WSServer transformer %T", logRec)
}

// ------------------------------------

type masmTransformer struct {
	t *masm.Transformer
}

// Transform transforms masm app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *masmTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*masm.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by masm transformer %T", logRec)
}

// ------------------------------------

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(appType string, version string, userMap *users.UserMap) (conversion.LogItemTransformer, error) {
	switch appType {
	case conversion.AppTypeAkalex, conversion.AppTypeCalc, conversion.AppTypeLists,
		conversion.AppTypeQuitaUp, conversion.AppTypeGramatikat:
		return &shinyTransformer{t: shiny.NewTransformer()}, nil
	case conversion.AppTypeKontext, conversion.AppTypeKontextAPI:
		switch version {
		case "0.13", "0.14":
			return &konText013Transformer{t: &kontext013.Transformer{}}, nil
		case "0.15", "0.16", "0.17":
			return &konText015Transformer{t: &kontext015.Transformer{}}, nil
		case "0.18":
			return &konText018Transformer{t: &kontext018.Transformer{}}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case conversion.AppTypeKwords:
		return &kwordsTransformer{t: &kwords.Transformer{}}, nil
	case conversion.AppTypeKorpusDB:
		return &korpusDBTransformer{t: &korpusdb.Transformer{}}, nil
	case conversion.AppTypeMapka:
		switch version {
		case "1":
			return &mapkaTransformer{t: mapka.NewTransformer()}, nil
		case "2":
			return &mapka2Transformer{t: mapka2.NewTransformer()}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case conversion.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{}}, nil
	case conversion.AppTypeSke:
		return &skeTransformer{t: ske.NewTransformer(userMap)}, nil
	case conversion.AppTypeSyd:
		return &sydTransformer{t: syd.NewTransformer(version)}, nil
	case conversion.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{}}, nil
	case conversion.AppTypeWag:
		switch version {
		case "0.6":
			return &wag06Transformer{t: &wag06.Transformer{}}, nil
		case "0.7":
			return &wag07Transformer{t: &wag07.Transformer{}}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case conversion.AppTypeWsserver:
		return &wsserverTransformer{t: &wsserver.Transformer{}}, nil
	case conversion.AppTypeMasm:
		return &masmTransformer{t: &masm.Transformer{}}, nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
