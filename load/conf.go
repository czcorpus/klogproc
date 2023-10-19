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

type ClusteringDBScanConf struct {
	MinDensity int     `json:"minDensity"`
	Epsilon    float64 `json:"epsilon"`
}

type BotDetectionConf struct {
	// IPOutlierCoeff specifies how far from the Q3 must a value be
	// to be considered an outlier (the formula is `Q3 + ipOutlierCoeff * IQR`)
	IPOutlierCoeff float64 `json:"ipOutlierCoeff"`

	// IPOutlierMinFreq specifies minimum number of requests for
	// an IP (per interval defined in buffer config `AnalysisIntervalSecs`)
	// to be actually reported. Because in case there is small traffic, even
	// legit IP requests may be evaluated as outliers.
	IPOutlierMinFreq int `json:"ipOutlierMinFreq"`

	// BlocklistIP is just for "known" IPs reporting (i.e. there is no
	// actual blocking involved - klogproc indeed cannot block anything).
	BlocklistIP []string `json:"blocklistIp"`

	// TrafficReportingThreshold defines a number specifying how much
	// a number of requests must have changed from the last check
	// (see `AnalysisIntervalSecs`) to be considered abnormal.
	// Please note that this number is really hard to tune as during
	// day, there are natural increases of traffic and without knowing
	// a typical (or even current) day requests progression, this is
	// rather a hint then a 100% evidence of bot activity.
	TrafficReportingThreshold float64 `json:"trafficReportingThreshold"`
}

type BufferConf struct {

	// ID buffers with ID can be shared between multiple log readers.
	// This makes sense mostly for services composed of multiple
	// homogenous processes each writing to its log file (e.g. Node.JS)
	ID string `json:"id"`

	HistoryLookupItems int `json:"historyLookupItems"`

	// AnalysisIntervalSecs specifies how often klogproc analyses previous
	// records. The interval is also important because it is a base for other
	// configured values (typically different limits/thresholds)
	AnalysisIntervalSecs int                   `json:"analysisIntervalSecs"`
	ClusteringDBScan     *ClusteringDBScanConf `json:"clusteringDbScan"`
	BotDetection         *BotDetectionConf     `json:"botDetection"`
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
