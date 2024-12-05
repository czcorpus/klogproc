// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package wag06

import (
	"klogproc/servicelog"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

func isQuery(action string) bool {
	return action == actionSearch || action == actionCompare || action == actionTranslate
}

func domainFromURL(urlAddr string) string {
	u, err := url.Parse(urlAddr)
	if err != nil {
		log.Warn().Msgf("Failed to parse url: %s", err)
		return ""
	}
	return u.Hostname()
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	ExcludeIPList servicelog.ExcludeIPList
}

func (t *Transformer) AppType() string {
	return servicelog.AppTypeWag
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(
	logRecord servicelog.InputRecord,
	tzShiftMin int,
) (servicelog.OutputRecord, error) {
	tLogRecord, ok := logRecord.(*InputRecord)
	if !ok {
		panic(servicelog.ErrFailedTypeAssertion)
	}
	r := &OutputRecord{
		Type:                t.AppType(),
		time:                tLogRecord.GetTime(),
		Datetime:            tLogRecord.GetTime().Add(time.Minute * time.Duration(tzShiftMin)).Format(time.RFC3339),
		IPAddress:           tLogRecord.Request.RemoteAddr,
		UserAgent:           tLogRecord.Request.HTTPUserAgent,
		ReferringDomain:     domainFromURL(tLogRecord.Request.Referer),
		IsAnonymous:         true, // from a web access log, we cannot extract the information
		IsQuery:             isQuery(tLogRecord.Action),
		IsMobileClient:      tLogRecord.IsMobileClient,
		HasPosSpecification: tLogRecord.HasPosSpecification,
		QueryType:           tLogRecord.QueryType,
		Lang1:               tLogRecord.Lang1,
		Lang2:               tLogRecord.Lang2,
		Queries:             tLogRecord.Queries,
		Action:              tLogRecord.Action,
		ProcTime:            tLogRecord.ProcTime,
	}
	r.ID = r.GenerateDeterministicID()
	return r, nil
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
