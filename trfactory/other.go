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

package trfactory

import (
	"fmt"
	"klogproc/logbuffer"
	"klogproc/servicelog"
	"klogproc/servicelog/korpusdb"
	"klogproc/servicelog/kwords"
	"klogproc/servicelog/morfio"
	"klogproc/servicelog/shiny"
	"klogproc/servicelog/ske"
	"klogproc/servicelog/syd"
	"klogproc/servicelog/treq"
)

// ------------------------------------

// ------------------------------------

type kwordsTransformer struct {
	t *kwords.Transformer
}

// Transform transforms KWords app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *kwordsTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*kwords.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by KWords transformer %T", logRec)
}

func (k *kwordsTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *kwordsTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type korpusDBTransformer struct {
	t *korpusdb.Transformer
}

// Transform transforms KorpusDB app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *korpusDBTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*korpusdb.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by KonText transformer %T", logRec)
}

func (k *korpusDBTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *korpusDBTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type morfioTransformer struct {
	t *morfio.Transformer
}

// Transform transforms Morfio app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *morfioTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*morfio.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by Morfio transformer %T", logRec)
}

func (k *morfioTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *morfioTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type skeTransformer struct {
	t *ske.Transformer
}

// Transform transforms SkE app log record (= web access log) types as general InputRecord
// In case of type mismatch, error is returned.
func (s *skeTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*ske.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by SkE transformer %T", logRec)
}

func (k *skeTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *skeTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type shinyTransformer struct {
	t *shiny.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *shinyTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*shiny.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by Shiny transformer %T", logRec)
}

func (k *shinyTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *shinyTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type sydTransformer struct {
	t *syd.Transformer
}

// Transform transforms SyD app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *sydTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*syd.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by SyD transformer %T", logRec)
}

func (k *sydTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *sydTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type treqTransformer struct {
	t *treq.Transformer
}

// Transform transforms Treq app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *treqTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*treq.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by Treq transformer %T", logRec)
}

func (k *treqTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *treqTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
