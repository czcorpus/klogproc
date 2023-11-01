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
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/fs"
	"github.com/rs/zerolog/log"
)

type Storable interface {
	GetTime() time.Time
	ClusteringClientID() string
}

type SerializableState interface {
	ToJSON() ([]byte, error)

	// AfterLoadNormalize should make sure loaded
	// data matches the provided `conf`.
	// E.g. if stored samples have length of 100
	// and the current configuration requires 20,
	// the method should cut the sample so it matches
	// the configuration.
	// It should also fix broken stored data (e.g. samples
	// with size 0)
	AfterLoadNormalize(conf *load.BufferConf, dt time.Time)

	// Report is mainly for debugging and overview
	// pursposes. It should show relevant values of the
	// state object.
	Report() map[string]any
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
	conf *load.BufferConf

	storageDirPath string

	// logFilePath refers to the log file that this buffer assists in processing
	logFilePath string

	data     map[string]*collections.CircularList[T]
	dataLock sync.RWMutex

	lastChecks     map[string]time.Time
	lastChecksLock sync.RWMutex

	stateData U

	hasLoadedStateData bool

	stateDataFactory func() U

	stateWriting chan U
}

func (st *Storage[T, U]) AddRecord(rec T) {
	st.dataLock.Lock()
	defer st.dataLock.Unlock()
	if st.conf.HistoryLookupItems > 0 {
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

func (st *Storage[T, U]) ClearOldRecords(maxAge time.Time) int {
	st.dataLock.Lock()
	defer st.dataLock.Unlock()
	var totalRm int
	for _, records := range st.data {
		records.ShiftUntil(func(item T) bool {
			ans := maxAge.After(item.GetTime())
			if ans {
				totalRm++
			}
			return ans
		})
	}
	return totalRm
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

// NewStorage is a recommended factory for creating `Storage`
func NewStorage[T Storable, U SerializableState](
	bufferConf *load.BufferConf,
	worklogReset bool,
	storageDirPath string,
	analyzedLogFilePath string,
	stateDataFactory func() U,
) *Storage[T, U] {
	if storageDirPath == "" {
		panic("no path specified for buffer state storage")
	}
	ans := &Storage[T, U]{
		data:             make(map[string]*collections.CircularList[T]),
		conf:             bufferConf,
		lastChecks:       make(map[string]time.Time),
		storageDirPath:   storageDirPath,
		logFilePath:      analyzedLogFilePath,
		stateWriting:     make(chan U),
		stateDataFactory: stateDataFactory,
	}
	fullPath := filepath.Join(storageDirPath, ans.mkStorageFileName())
	isF, _ := fs.IsFile(fullPath)
	if worklogReset && isF {
		err := os.Remove(fullPath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to remove buffer status file")

		} else {
			log.Info().Str("file", fullPath).Msg("removed buffer status file")
		}
	}

	go func() {
		for stateData := range ans.stateWriting {
			data, err := json.Marshal(stateData)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal log buffer state data")
			}
			if err = os.WriteFile(fullPath, data, 0644); err != nil {
				log.Error().Err(err).Msg("failed to store log buffer state data")
			}
		}
	}()
	return ans
}
