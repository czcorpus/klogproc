// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package elastic

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// ESQuery represents a structured type encoding an ElasticSearch
// query. Typically, it is a nested structure.
type ESQuery interface {
	ToJSONQuery() ([]byte, error)
}

// ----------------- query --------------------------

type datetimeRangeExpr struct {
	From string `json:"gte"`
	To   string `json:"lt"`
}

type datetimeRangeQuery struct {
	Datetime datetimeRangeExpr `json:"datetime"`
}

type rangeObj struct {
	Range datetimeRangeQuery `json:"range"`
}

type iPAddressExpr struct {
	IPAddress string `json:"ipAddress"`
}

type iPAddressTermObj struct {
	Term iPAddressExpr `json:"term"`
}

type userAgentExpr struct {
	UserAgent string `json:"userAgent"`
}

type userAgentMatchObj struct {
	Match userAgentExpr `json:"match"`
}

type appTypeExpr struct {
	AppType string `json:"type"`
}

type appTypeMatchObj struct {
	Match appTypeExpr `json:"match"`
}

type actionExpr struct {
	Action string `json:"action"`
}

type actionMatchObj struct {
	Match actionExpr `json:"match"`
}

type clientFlagExpr struct {
	ClientFlag string `json:"clientFlag"`
}

type clientFlagMatchObj struct {
	Match clientFlagExpr `json:"match"`
}

type pathExpr struct {
	Path string `json:"path"`
}

type pathMatchObj struct {
	Match pathExpr `json:"match"`
}

type boolObj struct {
	Must []interface{} `json:"must"`
}

type query struct {
	Bool boolObj `json:"bool"`
}

type srchQuery struct {
	Query query `json:"query"`
	From  int   `json:"from"`
	Size  int   `json:"size"`
}

func (sq *srchQuery) ToJSONQuery() ([]byte, error) {
	return json.Marshal(sq)
}

type countQuery struct {
	Query query `json:"query"`
}

func (cq *countQuery) ToJSONQuery() ([]byte, error) {
	return json.Marshal(cq)
}

// ---------------------------------------------------

// CNKRecordMeta contains meta information for a record
// as required by ElastiSearch bulk insert
type CNKRecordMeta struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
	Type  string `json:"_type"`
}

// ESCNKRecordMeta is just a wrapper for CNKRecordMeta
// as used when importing data
type ESCNKRecordMeta struct {
	Index CNKRecordMeta `json:"index"`
}

// ToJSON serializes the record to JSON
func (ecrm *ESCNKRecordMeta) ToJSON() ([]byte, error) {
	return json.Marshal(ecrm)
}

// ----------------- scroll -------------------------

type scrollObj struct {
	Scroll   string `json:"scroll"`
	ScrollID string `json:"scroll_id"`
}

// ----------------- result -------------------------

// ResultHit represents an individual result
type ResultHit struct {
	Index  string      `json:"_index"`
	Type   string      `json:"_type"`
	ID     string      `json:"_id"`
	Score  float32     `json:"_score"`
	Source interface{} `json:"_source"`
}

// Hits is a "hits" part of the ElasticSearch query result object
type Hits struct {
	Total    int         `json:"total"`
	MaxScore float32     `json:"max_score"`
	Hits     []ResultHit `json:"hits"`
}

func (h Hits) Sampled(prob float64) Hits {
	ans := Hits{
		Total:    h.Total,
		MaxScore: h.MaxScore,
		Hits:     make([]ResultHit, 0, int(math.Ceil(prob*float64(len(h.Hits))))),
	}
	for _, h := range h.Hits {
		if rand.Float64() < prob {
			ans.Hits = append(ans.Hits, h)
		}
	}
	return ans
}

// Result represents an ElasticSearch query result object
type Result struct {
	ScrollID string      `json:"_scroll_id"`
	Took     int         `json:"took"`
	TimedOut bool        `json:"timed_out"`
	Shards   interface{} `json:"_shards"`
	Hits     Hits        `json:"hits"`
}

// Result represents an ElasticSearch query count object
type CountResult struct {
	Count  int         `json:"count"`
	Shards interface{} `json:"_shards"`
}

// NewEmptyResult returns a new result with Total = 0
func NewEmptyResult() Result {
	return Result{Hits: Hits{Total: 0}}
}

func createBoolQuery(filter DocFilter) boolObj {
	m := boolObj{Must: make([]interface{}, 0)}
	if filter.FromDate != "" && filter.ToDate != "" {
		dateInterval := datetimeRangeExpr{From: filter.FromDate, To: filter.ToDate}
		m.Must = append(m.Must, &rangeObj{Range: datetimeRangeQuery{Datetime: dateInterval}})
	}
	if filter.IPAddress != "" {
		ipAddrObj := iPAddressTermObj{Term: iPAddressExpr{IPAddress: filter.IPAddress}}
		m.Must = append(m.Must, ipAddrObj)
	}
	if filter.UserAgent != "" {
		userAgentObj := userAgentMatchObj{userAgentExpr{UserAgent: filter.UserAgent}}
		m.Must = append(m.Must, userAgentObj)
	}
	if filter.AppType != "" {
		appTypeObj := appTypeMatchObj{appTypeExpr{AppType: filter.AppType}}
		m.Must = append(m.Must, appTypeObj)
	}
	if filter.Action != "" {
		actionObj := actionMatchObj{actionExpr{Action: filter.Action}}
		m.Must = append(m.Must, actionObj)
	}
	if filter.ClientFlag != "" {
		actionObj := clientFlagMatchObj{clientFlagExpr{ClientFlag: filter.ClientFlag}}
		m.Must = append(m.Must, actionObj)
	}
	if filter.Path != "" {
		actionObj := pathMatchObj{pathExpr{Path: filter.Path}}
		m.Must = append(m.Must, actionObj)
	}
	return m
}

// CreateClientSrchQuery generates a JSON-encoded query for ElastiSearch to
// find documents matching specified datetime range, optional IP
// address and optional userAgent substring/pattern
func CreateClientSrchQuery(filter DocFilter, chunkSize int) ([]byte, error) {
	if chunkSize < 1 {
		return []byte{}, fmt.Errorf("cannot load results of size < 1 (found %d)", chunkSize)
	}
	m := createBoolQuery(filter)
	q := srchQuery{Query: query{Bool: m}, From: 0, Size: chunkSize}
	return q.ToJSONQuery()
}

func CreateCountQuery(filter DocFilter) ([]byte, error) {
	m := createBoolQuery(filter)
	q := countQuery{Query: query{Bool: m}}
	return q.ToJSONQuery()
}

func createLogRecUpdQuery(upd DocUpdRecord) ([]byte, error) {
	d := docUpdObj{Doc: upd}
	return d.ToJSONQuery()
}

func createLogRecKeyRemoveQuery(key string) ([]byte, error) {
	d := docKeyRemoveObj{Script: fmt.Sprintf("ctx._source.remove('%s')", key)}
	return d.ToJSONQuery()
}
