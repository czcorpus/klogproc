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
	"klogproc/servicelog"
	"klogproc/servicelog/wag06"
	"klogproc/servicelog/wag07"
)

type wag06Transformer struct {
	t *wag06.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wag06Transformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*wag06.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by WaG 0.6 transformer %T", logRec)
}

func (k *wag06Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *wag06Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}

// ------------------------------------

type wag07Transformer struct {
	t *wag07.Transformer
}

// Transform transforms WaG app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *wag07Transformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*wag07.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by WaG 0.7 transformer %T", logRec)
}

func (k *wag07Transformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *wag07Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
