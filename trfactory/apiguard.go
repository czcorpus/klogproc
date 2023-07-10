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
	"klogproc/conversion/apiguard"
)

type apiguardTransformer struct {
	t *apiguard.Transformer
}

// Transform transforms APIGuard app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *apiguardTransformer) Transform(logRec conversion.InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*apiguard.InputRecord)
	if ok {
		return k.t.Transform(tRec, recType, tzShiftMin, anonymousUsers)
	}
	return nil, fmt.Errorf("invalid type for conversion by APIGuard transformer %T", logRec)
}

func (k *apiguardTransformer) HistoryLookupSecs() int {
	return k.t.HistoryLookupSecs()
}

func (k *apiguardTransformer) Preprocess(
	rec conversion.InputRecord, prevRecs []conversion.InputRecord,
) conversion.InputRecord {
	return k.t.Preprocess(rec, prevRecs)
}
