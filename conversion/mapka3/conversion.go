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

package mapka3

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"klogproc/clustering"
	"klogproc/conversion"
	"klogproc/load"
	"klogproc/logbuffer"

	"github.com/rs/zerolog/log"
)

// createID creates an idempotent ID of rec based on its properties.
func createID(rec *OutputRecord) string {
	str := rec.Type + rec.Path + rec.Datetime + rec.IPAddress + rec.UserID
	sum := sha1.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	bufferConf *load.BufferConf
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord *InputRecord,
	recType string,
	tzShiftMin int,
	anonymousUsers []int,
) (*OutputRecord, error) {

	r := &OutputRecord{
		Type:      recType,
		time:      logRecord.GetTime(),
		Datetime:  logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress: logRecord.Extra.IP,
		UserAgent: logRecord.GetUserAgent(),
		IsAnonymous: logRecord.Extra.UserID == "" ||
			conversion.UserBelongsToList(logRecord.Extra.UserID, anonymousUsers),
		Action:      "interaction",
		UserID:      logRecord.Extra.UserID,
		ClusterSize: logRecord.clusterSize,
	}
	if r.ClusterSize > 0 {
		r.IsQuery = true
	}
	r.ID = createID(r)
	return r, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return t.bufferConf.HistoryLookupItems
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord,
	prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	clusteringID := rec.ClusteringClientID()
	lastCheck := prevRecs.GetLastCheck(clusteringID)
	ci := time.Duration(t.bufferConf.AnalysisIntervalSecs) * time.Second
	if rec.GetTime().Sub(lastCheck) > ci {
		items := make([]conversion.InputRecord, 0, prevRecs.NumOfRecords(clusteringID))
		prevRecs.ForEach(clusteringID, func(item conversion.InputRecord) {
			items = append(items, item)
		})
		clustered := clustering.Analyze(
			t.bufferConf.ClusteringDBScan.MinDensity,
			t.bufferConf.ClusteringDBScan.Epsilon,
			items,
		)
		prevRecs.RemoveAnalyzedRecords(clusteringID, rec.GetTime())
		prevRecs.ConfirmRecordCheck(rec)
		log.Debug().
			Int("minDensity", t.bufferConf.ClusteringDBScan.MinDensity).
			Float64("epsilon", t.bufferConf.ClusteringDBScan.Epsilon).
			Time("searchStart", lastCheck).
			Time("searchEnd", rec.GetTime()).
			Int("foundClusters", len(clustered)).
			Msgf("log clustering in mapka3")
		return clustered
	}
	return []conversion.InputRecord{rec}
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(bufferConf *load.BufferConf) *Transformer {
	return &Transformer{
		bufferConf: bufferConf,
	}
}
