// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package analysis

import (
	"klogproc/logbuffer"
	"klogproc/servicelog"
)

type NullAnalyzer[T AnalyzableRecord] struct {
	appType string
}

func (analyzer *NullAnalyzer[T]) Preprocess(
	rec servicelog.InputRecord,
	prevRecs logbuffer.AbstractStorage[servicelog.InputRecord, logbuffer.SerializableState],
) []servicelog.InputRecord {
	return []servicelog.InputRecord{rec}
}

func NewNullAnalyzer[T AnalyzableRecord](appType string) *NullAnalyzer[T] {
	return &NullAnalyzer[T]{
		appType: appType,
	}
}
