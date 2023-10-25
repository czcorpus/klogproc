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
	"klogproc/servicelog/masm"
)

type masmTransformer struct {
	t *masm.Transformer
}

// Transform transforms masm app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *masmTransformer) Transform(logRec servicelog.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (servicelog.OutputRecord, error) {
	tRec, ok := logRec.(*masm.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for servicelog.by masm transformer %T", logRec)
}

func (k *masmTransformer) HistoryLookupItems() int {
	return k.t.HistoryLookupItems()
}

func (k *masmTransformer) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord],
) []servicelog.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
