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
	"encoding/json"
	"klogproc/clustering"
	"klogproc/load"
	"klogproc/logbuffer"
	"klogproc/servicelog"
	"time"

	"github.com/rs/zerolog/log"
)

// SimpleAnalysisState is mainly for debugging and/or reporting
// purposes.
type SimpleAnalysisState struct {
	LastCheck      time.Time `json:"timestamp"`
	TotalProcessed int       `json:"totalProcessed"`
	TotalIgnored   int       `json:"totalIgnored"`
}

func (state *SimpleAnalysisState) ToJSON() ([]byte, error) {
	return json.Marshal(state)
}

func (state *SimpleAnalysisState) AfterLoadNormalize(conf *load.BufferConf, dt time.Time) {
	if state.LastCheck.IsZero() && state.TotalProcessed > 0 {
		state.LastCheck = dt
	}
}

func (state *SimpleAnalysisState) Report() map[string]any {
	ans := make(map[string]any)
	ans["lastCheck"] = state.LastCheck
	ans["totalProcessed"] = state.TotalProcessed
	ans["totalIgnored"] = state.TotalIgnored
	return ans
}

type ClusteringAnalyzer[T AnalyzableRecord] struct {
	appType       string
	realtimeClock bool
	conf          *load.BufferConf
}

func (analyzer *ClusteringAnalyzer[T]) Preprocess(
	rec servicelog.InputRecord,
	prevRecs logbuffer.AbstractRecentRecords[servicelog.InputRecord, logbuffer.SerializableState],
) []servicelog.InputRecord {

	currTime := rec.GetTime()
	if analyzer.realtimeClock {
		currTime = time.Now()
	}

	stateData := prevRecs.GetStateData(currTime)
	tState, knownState := stateData.(*SimpleAnalysisState) // other types are ok here but no action will be done
	if knownState {
		tState.LastCheck = currTime
		tState.TotalProcessed++
	}
	clusteringID := rec.ClusteringClientID()
	clusterLastCheck := prevRecs.GetLastCheck(clusteringID)
	ci := time.Duration(analyzer.conf.AnalysisIntervalSecs) * time.Second
	if rec.GetTime().Sub(clusterLastCheck) > ci {
		items := make([]servicelog.InputRecord, 0, prevRecs.NumOfRecords(clusteringID))
		prevRecs.ForEach(clusteringID, func(item servicelog.InputRecord) {
			items = append(items, item)
		})
		if len(items) > 0 {
			clustered := clustering.Analyze(
				analyzer.conf.ClusteringDBScan.MinDensity,
				analyzer.conf.ClusteringDBScan.Epsilon,
				items,
			)
			log.Debug().
				Int("minDensity", analyzer.conf.ClusteringDBScan.MinDensity).
				Float64("epsilon", analyzer.conf.ClusteringDBScan.Epsilon).
				Time("firstRecord", items[0].GetTime()).
				Time("lastRecord", items[len(items)-1].GetTime()).
				Int("numAnalyzedRecords", len(items)).
				Int("foundClusters", len(clustered)).
				Msgf("log clustering in mapka3")

			if len(clustered) > 0 {
				prevRecs.RemoveAnalyzedRecords(clusteringID, rec.GetTime())
				prevRecs.ConfirmRecordCheck(rec)
				return clustered
			}
		}
	}
	return []servicelog.InputRecord{rec}
}

func NewAnalyzer[T AnalyzableRecord](
	appType string,
	conf *load.BufferConf,
	realtimeClock bool,
) *ClusteringAnalyzer[T] {
	return &ClusteringAnalyzer[T]{
		appType:       appType,
		conf:          conf,
		realtimeClock: realtimeClock,
	}
}
