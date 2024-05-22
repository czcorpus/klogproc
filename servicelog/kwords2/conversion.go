// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package kwords2

import (
	"klogproc/servicelog"
	"math"
	"strconv"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/rs/zerolog/log"
)

func convertMultitypeInt(v any) int {
	switch tv := v.(type) {
	case int:
		return tv
	case float64:
		return int(math.Round(tv))
	case string:
		v, err := strconv.Atoi(tv)
		if err != nil {
			return -1
		}
		return v
	}
	return -1
}

// --

type Transformer struct {
	ExcludeIPList servicelog.ExcludeIPList
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) []servicelog.InputRecord {
	if t.ExcludeIPList.Excludes(rec) {
		return []servicelog.InputRecord{}
	}
	return []servicelog.InputRecord{rec}
}

func (t *Transformer) Transform(
	logRecord *InputRecord,
	recType string,
	tzShiftMin int,
	anonymousUsers []int,
) (*OutputRecord, error) {
	var args *Args
	if logRecord.Action == "keywords/POST" {
		args = &Args{
			Attrs:        logRecord.Body.Attrs,
			Level:        logRecord.Body.Level,
			EffectMetric: logRecord.Body.EffectMetric,
			MinFreq:      convertMultitypeInt(logRecord.Body.MinFreq),
			Percent:      convertMultitypeInt(logRecord.Body.Percent),
		}
	}
	userID := logRecord.Headers.XUserID
	if userID == "" {
		switch tUserID := logRecord.UserID.(type) {
		case int:
			userID = strconv.Itoa(tUserID)
		case string:
			userID = tUserID
		}
	}
	if userID == "-1" {
		userID = ""
	}
	var userIDAttr *string
	isAnonymous := true
	if userID != "" {
		userIDAttr = &userID
		v, err := strconv.Atoi(userID)
		if err != nil {
			log.Error().Err(err).Str("value", userID).Msg("failed to parse user ID entry")

		} else {
			isAnonymous = collections.SliceContains(anonymousUsers, v)
		}
	}
	r := &OutputRecord{
		Type:          recType,
		Action:        logRecord.Action,
		Corpus:        logRecord.Body.RefCorpus,
		TextCharCount: logRecord.Body.TextCharCount,
		TextWordCount: logRecord.Body.TextWordCount,
		TextLang:      logRecord.Body.Lang,
		Datetime:      logRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:     logRecord.GetClientIP().String(),
		IsAnonymous:   isAnonymous,
		IsQuery:       logRecord.IsQuery,
		UserAgent:     logRecord.Headers.UserAgent,
		UserID:        userIDAttr,
		Error:         logRecord.ExportError(),
		Args:          args,
		Version:       "2",
	}
	r.ID = createID(r)
	return r, nil
}
