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

type DummyRecentRecords[T Storable, U SerializableState] struct {
	stateDataFactory func() U
}

func (st *DummyRecentRecords[T, U]) AddRecord(rec T) {
}

func (st *DummyRecentRecords[T, U]) ConfirmRecordCheck(rec Storable) {
}

func (st *DummyRecentRecords[T, U]) GetLastCheck(clusteringID string) time.Time {
	return time.Time{}
}

func (st *DummyRecentRecords[T, U]) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {
}

func (st *DummyRecentRecords[T, U]) NumOfRecords(clusteringID string) int {
	return 0
}

func (st *DummyRecentRecords[T, U]) ClearOldRecords(maxAge time.Time) int {
	return 0
}

func (st *DummyRecentRecords[T, U]) TotalNumOfRecordsSince(dt time.Time) int {
	return 0
}

func (st *DummyRecentRecords[T, U]) ForEach(clusteringID string, fn func(item T)) {
}

func (st *DummyRecentRecords[T, U]) TotalForEach(fn func(item T)) {
}

func (st *DummyRecentRecords[T, U]) SetStateData(stateData U) {
}

func (st *DummyRecentRecords[T, U]) GetStateData(dtNow time.Time) U {
	var u U
	return u
}

func (st *DummyRecentRecords[T, U]) EmptyStateData() U {
	return st.stateDataFactory()
}

func (st *DummyRecentRecords[T, U]) Report() map[string]any {
	return map[string]any{}
}

func NewDummyStorage[T Storable, U SerializableState](stateDataFactory func() U) *DummyRecentRecords[T, U] {
	return &DummyRecentRecords[T, U]{
		stateDataFactory: stateDataFactory,
	}
}
