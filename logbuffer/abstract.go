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

import "time"

type AbstractStorage[T Storable, U SerializableState] interface {
	AddRecord(rec T)

	// ConfirmRecordCheck sets a time of last check
	// for the records with the same clustering ID
	// as `rec`.
	// Please note that this also sets the global
	// timestamp (the same as `SetTimestamp` does).
	ConfirmRecordCheck(rec Storable)

	// GetLastCheck returns time of the last time items
	// with specified `clusteringID` where checked.
	// Note: global timestamp (`SetTimestamp`, `GetTimestamp`)
	// has no effect on this
	GetLastCheck(clusteringID string) time.Time

	// SetTimestamp stores provided time value typically to be able
	// to determine some previous global action. The function
	// returns previous stored timestamp
	SetTimestamp(t time.Time) time.Time

	// GetTimestamp returns stored timestamp
	GetTimestamp() time.Time

	// RemoveAnalyzedRecords removes all the records with specified
	// `clusteringID` up until the defined time.
	RemoveAnalyzedRecords(clusteringID string, dt time.Time)

	NumOfRecords(clusteringID string) int

	// TotalNumOfRecordsSince returns number of records
	// for each clusteringID with its time greater or equal
	// to the `dt`.
	TotalNumOfRecordsSince(dt time.Time) int

	ForEach(clusteringID string, fn func(item T))

	// TotalForEach apply a provided function on all items
	// no matter what clusteringID they belong to
	TotalForEach(fn func(item T))

	// SetAuxNumber can be used to store custom data for later usage
	// (e.g. summaries of previous 'Preprocess' calls)
	SetAuxNumber(name string, value float64)

	// GetAuxNumber returns previously stored custom float value
	GetAuxNumber(name string) (float64, bool)

	// AddNumberSample adds a new value to the "sample storage".
	// It returns the sample size after the value was added
	AddNumberSample(storageKey string, value float64) int

	GetNumberSamples(storageKey string) []float64

	StoreStateData(stateData U) error

	LoadStateData() (U, error)
}
