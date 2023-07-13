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
	"klogproc/conversion"
	"klogproc/conversion/mapka"
	"klogproc/conversion/mapka2"
	"klogproc/conversion/mapka3"
	"klogproc/logbuffer"
)

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

func (k *mapkaTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *mapkaTransformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
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

func (k *mapka2Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *mapka2Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type mapka3Transformer struct {
	t *mapka3.Transformer
}

// Transform transforms Mapka (v2) app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *mapka3Transformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*mapka3.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by Mapka3 transformer %T", logRec)
}

func (k *mapka3Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *mapka3Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
