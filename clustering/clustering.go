// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package clustering

import (
	"klogproc/conversion"
	"time"

	"github.com/kelindar/dbscan"
)

type ClusterableRecord struct {
	rec conversion.InputRecord
}

func (cr ClusterableRecord) GetTime() time.Time {
	return cr.rec.GetTime()
}

func (cr ClusterableRecord) DistanceTo(other dbscan.Point) float64 {
	return other.(ClusterableRecord).GetTime().Sub(cr.rec.GetTime()).Seconds()
}

func (cr ClusterableRecord) Name() string {
	return cr.rec.GetTime().Format(time.RFC3339)
}

func wrapInputRecords(input []conversion.InputRecord) []dbscan.Point {
	ans := make([]dbscan.Point, len(input))
	for i, v := range input {
		ans[i] = ClusterableRecord{rec: v}
	}
	return ans
}

func Analyze(
	minDensity int, epsilon float64, input []conversion.InputRecord,
) []conversion.InputRecord {
	input2 := wrapInputRecords(input)
	clusters := dbscan.Cluster(minDensity, epsilon, input2...)
	ans := make([]conversion.InputRecord, len(clusters))
	for i, cl := range clusters {
		rec := (cl[0].(ClusterableRecord)).rec
		rec.SetCluster(len(cl))
		ans[i] = rec
	}
	return ans
}
