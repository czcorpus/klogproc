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
	"github.com/czcorpus/klogproc/conversion/calc"
	"github.com/czcorpus/klogproc/conversion/kontext"
	"github.com/czcorpus/klogproc/conversion/kwords"
	"github.com/czcorpus/klogproc/conversion/morfio"
	"github.com/czcorpus/klogproc/conversion/ske"
	"github.com/czcorpus/klogproc/conversion/syd"
	"github.com/czcorpus/klogproc/conversion/treq"
	"github.com/czcorpus/klogproc/conversion/wag"
	"github.com/czcorpus/klogproc/users"
)

// ------------------------------------

// konTextTransformer is a simple type-safe wrapper for kontext app type log transformer
type konTextTransformer struct {
	t *kontext.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *konTextTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kontext.InputRecord)
	if ok {
		return k.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

type sydTransformer struct {
	t *syd.Transformer
}

// Transform transforms SyD app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *sydTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*syd.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by SyD transformer %T", logRec)
}

// ------------------------------------

type treqTransformer struct {
	t *treq.Transformer
}

// Transform transforms Treq app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *treqTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*treq.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Treq transformer %T", logRec)
}

// ------------------------------------

type morfioTransformer struct {
	t *morfio.Transformer
}

// Transform transforms Morfio app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *morfioTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*morfio.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Morfio transformer %T", logRec)
}

// ------------------------------------

type kwordsTransformer struct {
	t *kwords.Transformer
}

// Transform transforms KWords app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *kwordsTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kwords.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KWords transformer %T", logRec)
}

// ------------------------------------

type skeTransformer struct {
	t *ske.Transformer
}

// Transform transforms SkE app log record (= web access log) types as general InputRecord
// In case of type mismatch, error is returned.
func (s *skeTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*ske.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by SkE transformer %T", logRec)
}

// ------------------------------------

type wagTransformer struct {
	t *wag.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wagTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*wag.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by WaG transformer %T", logRec)
}

// ------------------------------------

type calcTransformer struct {
	t *calc.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *calcTransformer) Transform(logRec conversion.InputRecord, recType string, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*calc.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, anonymousUsers)
	}
	return nil, fmt.Errorf("Invalid type for conversion by Calc transformer %T", logRec)
}

// ------------------------------------

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(appType string, version int, userMap *users.UserMap) (conversion.LogItemTransformer, error) {

	switch appType {
	case conversion.AppTypeKontext:
		return &konTextTransformer{t: &kontext.Transformer{}}, nil
	case conversion.AppTypeSyd:
		return &sydTransformer{t: syd.NewTransformer(version)}, nil
	case conversion.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{}}, nil
	case conversion.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{}}, nil
	case conversion.AppTypeKwords:
		return &kwordsTransformer{t: &kwords.Transformer{}}, nil
	case conversion.AppTypeSke:
		return &skeTransformer{t: ske.NewTransformer(userMap)}, nil
	case conversion.AppTypeWag:
		return &wagTransformer{t: &wag.Transformer{}}, nil
	case conversion.AppTypeCalc:
		return &calcTransformer{t: calc.NewTransformer()}, nil
	default:
		return nil, fmt.Errorf("Cannot find log transformer for app type %s", appType)
	}
}
