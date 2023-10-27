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
	ConfirmRecordCheck(rec Storable)

	// GetLastCheck returns time of the last time items
	// with specified `clusteringID` where checked.
	GetLastCheck(clusteringID string) time.Time

	// RemoveAnalyzedRecords removes all the records with specified
	// `clusteringID` up until the defined time.
	RemoveAnalyzedRecords(clusteringID string, dt time.Time)

	NumOfRecords(clusteringID string) int

	// ClearOldRecords is a maintenance function called
	// randomly by a respective log processor to keep
	// the number of records in RAM at a reasonable level.
	// The method should return number of removed items
	// (it is mostly for better overview, i.e. not essential)
	ClearOldRecords(maxAge time.Time) int

	// TotalNumOfRecordsSince returns number of records
	// for each clusteringID with its time greater or equal
	// to the `dt`.
	TotalNumOfRecordsSince(dt time.Time) int

	ForEach(clusteringID string, fn func(item T))

	// TotalForEach apply a provided function on all items
	// no matter what clusteringID they belong to
	TotalForEach(fn func(item T))

	// SetStateData sets the current state data for later reuse
	// It may or may not backup data to disk or database.
	// If applicable then `GetStateData` should load the data
	// in case nothing was set yet.
	SetStateData(stateData U)

	GetStateData() U

	EmptyStateData() U
}
