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
	"encoding/json"
	"klogproc/load"
	"sync"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
)

type Storable interface {
	GetTime() time.Time
	ClusteringClientID() string
}

type SerializableState interface {
	json.Marshaler
}

// Storage keeps a defined number of log records in memory
// (using a circular list) for log processors which need not
// just the current log line/record but also some history
// lookup (e.g. for record clustering in `mapka` or for
// bot detection)
// Besides stored records, it provides a simple interface
// for storing and retrieving misc. values for log processors
// to be able to evaluate recent state.
// All the functions are safe to be used concurrently.
// The `T` type represents a log record type stored by this Storage.
// The `U` type is a type used to store state data when dealing
// with persistence.
type Storage[T Storable, U SerializableState] struct {
	initialCapacity int

	storageDirPath string

	// logFilePath refers to the log file that this buffer assists in processing
	logFilePath string

	data     map[string]*collections.CircularList[T]
	dataLock sync.RWMutex

	lastChecks     map[string]time.Time
	lastChecksLock sync.RWMutex

	// auxNumbers can be used to store some auxiliary summaries
	auxNumbers     map[string]float64
	auxNumbersLock sync.RWMutex

	auxNumberSamples     map[string]*SampleWithReplac[float64]
	auxNumberSamplesLock sync.RWMutex

	timestamp time.Time
}

func (st *Storage[T, U]) AddRecord(rec T) {
	st.dataLock.Lock()
	defer st.dataLock.Unlock()
	if st.initialCapacity > 0 {
		cid := rec.ClusteringClientID()
		_, ok := st.data[cid]
		if !ok {
			st.data[cid] = collections.NewCircularList[T](1000)
		}
		st.data[cid].Append(rec)
	}
}

func (st *Storage[T, U]) ConfirmRecordCheck(rec Storable) {
	st.lastChecksLock.Lock()
	defer st.lastChecksLock.Unlock()
	st.lastChecks[rec.ClusteringClientID()] = rec.GetTime()
}

func (st *Storage[T, U]) GetLastCheck(clusteringID string) time.Time {
	st.lastChecksLock.RLock()
	defer st.lastChecksLock.RUnlock()
	v := st.lastChecks[clusteringID]
	return v
}

// SetTimestamp sets a global (for a concrete log file processing)
// timestamp. This is typically used to mark last log analysis
// when detecting bots or errors.
func (st *Storage[T, U]) SetTimestamp(t time.Time) time.Time {
	prev := st.timestamp
	st.timestamp = t
	return prev
}

// GetTimestamp gets a global (for a concrete log file processing)
// timestamp. This is typically used to mark last log analysis
// when detecting bots or errors.
func (st *Storage[T, U]) GetTimestamp() time.Time {
	return st.timestamp
}

// RemoveAnalyzedRecords removes all the log records older than `dt`
// with provided `clusteringID` (which is typically something like userID, session, IP)
func (st *Storage[T, U]) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {
	st.dataLock.Lock()
	defer st.dataLock.Unlock()
	v, ok := st.data[clusteringID]
	if !ok {
		return
	}
	v.ShiftUntil(func(item T) bool {
		return item.GetTime().Before(dt)
	})
}

// NumOfRecords gets number of stored records for a specific
// records (identified by their `clusteringID`).
func (st *Storage[T, U]) NumOfRecords(clusteringID string) int {
	st.dataLock.RLock()
	defer st.dataLock.RUnlock()
	v, ok := st.data[clusteringID]
	if !ok {
		return 0
	}
	return v.Len()
}

// TotalNumOfRecords returns total number of stored records
// no matter what clustering ID they have but with its
// time greater or equal to `dt`
func (st *Storage[T, U]) TotalNumOfRecordsSince(dt time.Time) int {
	st.dataLock.RLock()
	defer st.dataLock.RUnlock()
	var ans int
	for _, v := range st.data {
		v.ForEach(func(i int, item T) bool {
			if item.GetTime().After(dt) || item.GetTime().Equal(dt) {
				ans++
			}
			return true
		})
	}
	return ans
}

// ForEach iterates over stored records with the provided `clusteringID`
// and calls the provided `fn` with each item as an argument.
func (st *Storage[T, U]) ForEach(clusteringID string, fn func(item T)) {
	st.dataLock.RLock()
	defer st.dataLock.RUnlock()
	v, ok := st.data[clusteringID]
	if !ok {
		return
	}
	v.ForEach(func(i int, item T) bool {
		fn(item)
		return true
	})
}

// ForEach iterates over all stored records (no matter what clustering ID they have)
// and calls the provided `fn` with each item as an argument.
//
// Please note that the records are not sorted by date here as the method
// iterates in two nested loops - first one goes through all the record groups
// (= records with the same clustering ID) and the for each this group it iterates
// through all its items.
func (st *Storage[T, U]) TotalForEach(fn func(item T)) {
	st.dataLock.RLock()
	defer st.dataLock.RUnlock()
	for _, v := range st.data {
		v.ForEach(func(i int, item T) bool {
			fn(item)
			return true
		})
	}
}

// SetAuxNumber sets an auxiliary number for later reuse.
func (st *Storage[T, U]) SetAuxNumber(name string, value float64) {
	st.auxNumbersLock.Lock()
	defer st.auxNumbersLock.Unlock()
	st.auxNumbers[name] = value
}

// GetAuxNumber gets a previously stored auxiliary number.
func (st *Storage[T, U]) GetAuxNumber(name string) (float64, bool) {
	st.auxNumbersLock.RLock()
	defer st.auxNumbersLock.RUnlock()
	v, ok := st.auxNumbers[name]
	return v, ok
}

// AddNumberSample adds a new number to a "sample pool" which is
// basically a list of numbers of a fixed size (= historyLookupItems).
// When the list is full, items are replaced randomly by incoming values.
//
// This is mostly for keeping track of the current traffic intensity
func (st *Storage[T, U]) AddNumberSample(storageKey string, value float64) int {
	st.auxNumberSamplesLock.Lock()
	defer st.auxNumberSamplesLock.Unlock()
	samples, ok := st.auxNumberSamples[storageKey]
	if !ok {
		samples = NewSampleWithReplac[float64](st.initialCapacity)
		st.auxNumberSamples[storageKey] = samples
	}
	return samples.Add(value)
}

// GetNumberSamples returns a list of previously stored numbers sample.
// See `AddNumberSample` for more info.
func (st *Storage[T, U]) GetNumberSamples(storageKey string) []float64 {
	st.auxNumberSamplesLock.RLock()
	defer st.auxNumberSamplesLock.RUnlock()
	samples, ok := st.auxNumberSamples[storageKey]
	if !ok {
		return []float64{}
	}
	return samples.GetAll()
}

// NewStorage is a recommended factory for creating `Storage`
func NewStorage[T Storable, U SerializableState](
	bufferConf *load.BufferConf,
	storageDirPath string,
	analyzedLogFilePath string,
) *Storage[T, U] {
	return &Storage[T, U]{
		data:             make(map[string]*collections.CircularList[T]),
		initialCapacity:  bufferConf.HistoryLookupItems,
		lastChecks:       make(map[string]time.Time),
		auxNumbers:       make(map[string]float64),
		auxNumberSamples: make(map[string]*SampleWithReplac[float64]),
		storageDirPath:   storageDirPath,
		logFilePath:      analyzedLogFilePath,
	}
}
