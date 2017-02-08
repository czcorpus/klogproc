// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

import "encoding/json"

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

type boolObj struct {
	Must []interface{} `json:"must"`
}

type query struct {
	Bool boolObj `json:"bool"`
}

type srchQuery struct {
	Query query `json:"query"`
}

func (sq *srchQuery) ToJSONQuery() ([]byte, error) {
	return json.Marshal(sq)
}

// ---------------------------------------------------

// APIFlagUpdateConf specifies parameters of "isApi" flag operation
type APIFlagUpdateConf struct {
	Disabled  bool   `json:"disabled"`
	FromDate  string `json:"fromDate"`
	ToDate    string `json:"toDate"`
	IPAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}

// ------------------ record update -------------------

type docUpdObj struct {
	Doc docRecord `json:"doc"`
}

func (duo *docUpdObj) ToJSONQuery() ([]byte, error) {
	return json.Marshal(duo)
}

type docRecord struct {
	IsAPI bool `json:"isAPI"`
}

type UpdResponse struct {
	Index   string      `json:"_index"`
	Type    string      `json:"_type"`
	ID      string      `json:"_id"`
	Version int         `json:"_version"`
	Result  string      `json:"result"`
	Shards  interface{} `json:"_shards"` // we don't care much about this (yet)
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

// Result represents an ElasticSearch query result object
type Result struct {
	Took     int         `json:"took"`
	TimedOut bool        `json:"timed_out"`
	Shards   interface{} `json:"_shards"`
	Hits     Hits        `json:"hits"`
}

// NewEmptyResult returns a new result with Total = 0
func NewEmptyResult() Result {
	return Result{Hits: Hits{Total: 0}}
}

// CreateClientSrchQuery generates a JSON-encoded query for ElastiSearch to
// find documents matching specified datetime range, optional IP
// address and optional userAgent substring/pattern
func CreateClientSrchQuery(fromDate string, toDate string, ipAddress string, userAgent string) ([]byte, error) {
	m := boolObj{Must: make([]interface{}, 1)}
	dateInterval := datetimeRangeExpr{From: fromDate, To: toDate}
	m.Must[0] = &rangeObj{Range: datetimeRangeQuery{Datetime: dateInterval}}
	if ipAddress != "" {
		ipAddrObj := iPAddressTermObj{Term: iPAddressExpr{IPAddress: ipAddress}}
		m.Must = append(m.Must, ipAddrObj)
	}
	if userAgent != "" {
		userAgentObj := userAgentMatchObj{userAgentExpr{UserAgent: userAgent}}
		m.Must = append(m.Must, userAgentObj)
	}
	q := srchQuery{Query: query{Bool: m}}
	return q.ToJSONQuery()
}

func CreateClientAPIFlagUpdQuery() ([]byte, error) {
	d := docUpdObj{Doc: docRecord{IsAPI: true}}
	return d.ToJSONQuery()
}
