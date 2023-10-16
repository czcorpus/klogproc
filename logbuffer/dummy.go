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

package logbuffer

import (
	"time"
)

type DummyStorage[T Storable] struct {
}

func (st *DummyStorage[T]) AddRecord(rec T) {
}

func (st *DummyStorage[T]) ConfirmRecordCheck(rec Storable) {
}

func (st *DummyStorage[T]) GetLastCheck(clusteringID string) time.Time {
	return time.Time{}
}

func (st *DummyStorage[T]) SetTimestamp(t time.Time) time.Time {
	return time.Time{}
}

func (st *DummyStorage[T]) GetTimestamp() time.Time {
	return time.Time{}
}

func (st *DummyStorage[T]) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {
}

func (st *DummyStorage[T]) TotalRemoveAnalyzedRecords(dt time.Time) {
}

func (st *DummyStorage[T]) NumOfRecords(clusteringID string) int {
	return 0
}

func (st *DummyStorage[T]) TotalNumOfRecords() int {
	return 0
}

func (st *DummyStorage[T]) ForEach(clusteringID string, fn func(item T)) {
}

func (st *DummyStorage[T]) TotalForEach(fn func(item T)) {
}

func (st *DummyStorage[T]) SetAuxNumber(name string, value float64) {
}

func (st *DummyStorage[T]) GetAuxNumber(name string) (float64, bool) {
	return 0, false
}

func (st *DummyStorage[T]) AddNumberSample(storageKey string, value float64) int {
	return 0
}

func (st *DummyStorage[T]) GetNumberSamples(storageKey string) []float64 {
	return []float64{}
}

func NewDummyStorage[T Storable]() *DummyStorage[T] {
	return &DummyStorage[T]{}
}