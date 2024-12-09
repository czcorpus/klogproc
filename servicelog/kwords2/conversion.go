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
	AnonymousUsers []int
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeKwords
}

func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	var args *Args
	if tLogRecord.Action == "keywords/POST" {
		args = &Args{
			Attrs:        tLogRecord.Body.Attrs,
			Level:        tLogRecord.Body.Level,
			EffectMetric: tLogRecord.Body.EffectMetric,
			MinFreq:      convertMultitypeInt(tLogRecord.Body.MinFreq),
			Percent:      convertMultitypeInt(tLogRecord.Body.Percent),
		}
	}
	userID := tLogRecord.Headers.XUserID()
	if userID == "" {
		switch tUserID := tLogRecord.UserID.(type) {
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
			isAnonymous = collections.SliceContains(t.AnonymousUsers, v)
		}
	}
	r := &OutputRecord{
		Type:          t.AppType(),
		Action:        tLogRecord.Action,
		Corpus:        tLogRecord.Body.RefCorpus,
		TextCharCount: tLogRecord.Body.TextCharCount,
		TextWordCount: tLogRecord.Body.TextWordCount,
		TextLang:      tLogRecord.Body.Lang,
		Datetime:      tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:     tLogRecord.GetClientIP().String(),
		IsAnonymous:   isAnonymous,
		IsQuery:       tLogRecord.IsQuery,
		UserAgent:     tLogRecord.Headers.UserAgent(),
		UserID:        userIDAttr,
		Error:         tLogRecord.ExportError(),
		Args:          args,
		Version:       servicelog.AppVersionKwords2,
	}
	r.ID = r.GenerateDeterministicID()
	return r, nil
}
