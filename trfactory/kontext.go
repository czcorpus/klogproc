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
	"klogproc/conversion/kontext013"
	"klogproc/conversion/kontext015"
	"klogproc/conversion/kontext018"
	"klogproc/logbuffer"
)

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

func (k *konText013Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *konText013Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
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

func (k *konText015Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *konText015Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
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

func (k *konText018Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *konText018Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
