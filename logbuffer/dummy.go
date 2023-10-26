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

type DummyStorage[T Storable, U SerializableState] struct {
}

func (st *DummyStorage[T, U]) AddRecord(rec T) {
}

func (st *DummyStorage[T, U]) ConfirmRecordCheck(rec Storable) {
}

func (st *DummyStorage[T, U]) GetLastCheck(clusteringID string) time.Time {
	return time.Time{}
}

func (st *DummyStorage[T, U]) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {
}

func (st *DummyStorage[T, U]) NumOfRecords(clusteringID string) int {
	return 0
}

func (st *DummyStorage[T, U]) TotalNumOfRecordsSince(dt time.Time) int {
	return 0
}

func (st *DummyStorage[T, U]) ForEach(clusteringID string, fn func(item T)) {
}

func (st *DummyStorage[T, U]) TotalForEach(fn func(item T)) {
}

func (st *DummyStorage[T, U]) SetStateData(stateData U) {
}

func (st *DummyStorage[T, U]) GetStateData() U {
	var u U
	return u
}

func (st *DummyStorage[T, U]) LoadStateData() (U, error) {
	var ans U
	return ans, nil
}

func NewDummyStorage[T Storable, U SerializableState]() *DummyStorage[T, U] {
	return &DummyStorage[T, U]{}
}
