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
	"klogproc/servicelog"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/maths"
)

type sitemsWrapper struct {
	data collections.BinTree[*ReqCalcItem]
}

func (w *sitemsWrapper) Get(idx int) maths.FreqInfo {
	return w.data.Get(idx)
}

func (w *sitemsWrapper) Len() int {
	return w.data.Len()
}

type ReqCalcItem struct {
	IP                   string
	Count                int
	Known                bool
	CountWholeBuffer     int
	NumSuspicWholeBuffer int
}

func (rci *ReqCalcItem) SuspicRatio() float64 {
	return float64(rci.NumSuspicWholeBuffer) / float64(rci.CountWholeBuffer)
}

type AnalyzableRecord interface {
	servicelog.InputRecord
	ShouldBeAnalyzed() bool
}

// Freq is implemented to satisfy cnc-gokit utils
func (rc *ReqCalcItem) Freq() int {
	return rc.Count
}

func (rc *ReqCalcItem) Compare(other collections.Comparable) int {
	if rc.Count > other.(*ReqCalcItem).Count {
		return 1

	} else if rc.Count == other.(*ReqCalcItem).Count {
		return 0
	}
	return -1
}
