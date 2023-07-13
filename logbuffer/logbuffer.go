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
	"klogproc/load"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
)

type Storable interface {
	GetTime() time.Time
	ClusteringClientID() string
}

type Storage[T Storable] struct {
	initialCapacity int
	data            map[string]*collections.CircularList[T]
	lastChecks      map[string]time.Time
}

func (st *Storage[T]) AddRecord(rec T) {
	if st.initialCapacity > 0 {
		cid := rec.ClusteringClientID()
		_, ok := st.data[cid]
		if !ok {
			st.data[cid] = collections.NewCircularList[T](1000)
		}
		st.data[cid].Append(rec)
	}
}

func (st *Storage[T]) ConfirmRecordCheck(rec Storable) {
	st.lastChecks[rec.ClusteringClientID()] = rec.GetTime()
}

func (st *Storage[T]) GetLastCheck(clusteringID string) time.Time {
	v := st.lastChecks[clusteringID]
	return v
}

func (st *Storage[T]) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {
	v, ok := st.data[clusteringID]
	if !ok {
		return
	}
	v.ShiftUntil(func(item T) bool {
		return item.GetTime().Before(dt)
	})
}

func (st *Storage[T]) NumOfRecords(clusteringID string) int {
	v, ok := st.data[clusteringID]
	if !ok {
		return 0
	}
	return v.Len()
}

func (st *Storage[T]) ForEach(clusteringID string, fn func(item T)) {
	v, ok := st.data[clusteringID]
	if !ok {
		return
	}
	v.ForEach(func(i int, item T) bool {
		fn(item)
		return true
	})
}

func NewStorage[T Storable](bufferConf *load.BufferConf) *Storage[T] {
	return &Storage[T]{
		data:            make(map[string]*collections.CircularList[T]),
		initialCapacity: bufferConf.HistoryLookupItems,
		lastChecks:      make(map[string]time.Time),
	}
}
