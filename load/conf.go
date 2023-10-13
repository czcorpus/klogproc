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

package load

import "errors"

type BufferConf struct {
	HistoryLookupItems   int `json:"historyLookupItems"`
	AnalysisIntervalSecs int `json:"analysisIntervalSecs"`
	ClusteringDBScan     *struct {
		MinDensity int     `json:"minDensity"`
		Epsilon    float64 `json:"epsilon"`
	} `json:"clusteringDbScan"`
	BotDetection *struct {
		IPOutlierCoeff float64  `json:"ipOutlierCoeff"`
		BlocklistIP    []string `json:"blocklistIp"`
	} `json:"botDetection"`
}

func (bc *BufferConf) Validate() error {
	if bc.HistoryLookupItems <= 0 {
		return errors.New(
			"failed to validate batch file processing buffer: historyLookupItems must be > 0")
	}
	if bc.AnalysisIntervalSecs <= 0 {
		return errors.New(
			"failed to validate batch file processing buffer: analysisIntervalSecs must be > 0")
	}
	if bc.ClusteringDBScan != nil {
		if bc.ClusteringDBScan.Epsilon <= 0 {
			return errors.New(
				"failed to validate batch file processing buffer: clusteringDbScan.epsilon must be > 0")
		}
		if bc.ClusteringDBScan.MinDensity <= 0 {
			return errors.New(
				"failed to validate batch file processing buffer: clusteringDbScan.minDensity must be > 0")
		}
	}
	return nil
}
