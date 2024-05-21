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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	defaultReqTimeoutSecs = 10
)

// ConnectionConf defines a configuration
// required to work with ES client.
type ConnectionConf struct {
	Server         string `json:"server"`
	Index          string `json:"index"`
	PushChunkSize  int    `json:"pushChunkSize"`
	ScrollTTL      string `json:"scrollTtl"`
	ReqTimeoutSecs int    `json:"reqTimeoutSecs"`
	MajorVersion   int    `json:"majorVersion"`
}

// IsConfigured tests whether the configuration is considered
// to be enabled (i.e. no error checking just enabled/disabled)
func (conf *ConnectionConf) IsConfigured() bool {
	return conf.Server != ""
}

// Validate tests whether the configuration is filled in
// correctly. Please note that if the function returns nil
// then IsConfigured() must return 'true'.
func (conf *ConnectionConf) Validate() error {
	if conf.Index == "" {
		return fmt.Errorf("ERROR: index/indexPrefix not set for ElasticSearch")
	}
	if conf.ScrollTTL == "" {
		return fmt.Errorf("ERROR: elasticScrollTtl must be a valid ElasticSearch scroll arg value (e.g. '2m', '30s')")
	}
	if conf.PushChunkSize == 0 {
		return fmt.Errorf("ERROR: elasticPushChunkSize is missing")
	}
	if conf.ReqTimeoutSecs == 0 {
		conf.ReqTimeoutSecs = defaultReqTimeoutSecs
		log.Warn().Msgf("missing elasticSearch.reqTimeoutSecs, using default %d", defaultReqTimeoutSecs)
	}
	return nil
}

// -------

type BulkWriteRespItem struct {
	Index struct {
		Error struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		}
		Status int `json:"status"`
	} `json:"index"`
}

// BulkWriteResp represents a response from an ElasticSearch
// server to a bulk insert action. Please note that even
// there is an error (or more errors), the return code is
// still 200 and actual individual errors are stored in
// 'Items'.
// TODO - test whether this works also for ES >= 6
type BulkWriteResp struct {
	Took   int                 `json:"took"`
	Errors bool                `json:"errors"`
	Items  []BulkWriteRespItem `json:"items"`
}

func (bwresp BulkWriteResp) FirstError() string {
	if len(bwresp.Items) > 0 && bwresp.Items[0].Index.Error.Type != "" {
		return fmt.Sprintf(
			"%s (%s)", bwresp.Items[0].Index.Error.Type, bwresp.Items[0].Index.Error.Reason)
	}
	return ""
}

// ErrorResultObj describes an error response from ElasticSearch
type ErrorResultObj struct {
	Error  map[string]interface{} `json:"error"`
	Status int                    `json:"status"`
}

func (ero ErrorResultObj) String() string {
	var ans bytes.Buffer
	for k, v := range ero.Error {
		ans.WriteString(fmt.Sprintf("{%s -> %s}", k, v))
	}
	return ans.String()
}

// ESClientError is a general response error
type ESClientError struct {
	error
	Query   []byte
	ESError ErrorResultObj
}

func (esc *ESClientError) Error() string {
	return fmt.Sprintf("%s: %s", esc.error.Error(), esc.ESError)
}

func newESClientError(message string, response []byte, query []byte) *ESClientError {
	var errResult ErrorResultObj
	json.Unmarshal(response, &errResult)
	return &ESClientError{errors.New(message), query, errResult}
}

// ESClient is a simple ElasticSearch client
type ESClient struct {
	server         string
	index          string
	reqTimeoutSecs int
	version        int
}

// NewClient returns an instance of ESClient
func NewClient(conf *ConnectionConf) *ESClient {
	return &ESClient{
		server:         conf.Server,
		index:          conf.Index,
		reqTimeoutSecs: conf.ReqTimeoutSecs,
		version:        5,
	}
}

// NewClient6 returns an instance of ESClient for ElasticSearch 6+
func NewClient6(conf *ConnectionConf, appType string) *ESClient {
	return &ESClient{
		server:         conf.Server,
		index:          fmt.Sprintf("%s_%s", conf.Index, appType),
		reqTimeoutSecs: conf.ReqTimeoutSecs,
		version:        6,
	}
}

func (c ESClient) String() string {
	return fmt.Sprintf(
		"ESClient{server: %s, index: %s, version: %d}",
		c.server, c.index, c.version)
}

// Do sends a general request to ElasticSearch server where
// 'query' is expected to be a JSON-encoded argument object
func (c *ESClient) Do(method string, path string, query []byte) ([]byte, error) {
	body := bytes.NewBuffer(query)
	client := http.Client{Timeout: time.Second * time.Duration(c.reqTimeoutSecs)}
	req, err := http.NewRequest(method, c.server+path, body)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return respBody, newESClientError(
			fmt.Sprintf("Request %s failed with code %d", path, resp.StatusCode), respBody, query)
	}

	var resObj BulkWriteResp
	err = json.Unmarshal(respBody, &resObj)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decode ES bulk write response: %w", err)
	}
	firstErr := resObj.FirstError()
	if firstErr != "" {
		return []byte{}, fmt.Errorf("failed to write data to ES: %w", errors.New(firstErr))
	}
	if err != nil {
		return []byte{}, err
	}
	return respBody, nil
}

// search is a low level search function
func (c *ESClient) search(query []byte, scroll string) (Result, error) {
	path := "/" + c.index + "/_search"
	if scroll != "" {
		path += "?scroll=" + scroll
	}
	resp, err := c.Do("GET", path, query)
	if err != nil {
		return NewEmptyResult(), err
	}
	var srchResult Result
	err2 := json.Unmarshal(resp, &srchResult)
	if err2 == nil {
		return srchResult, err2
	}
	return NewEmptyResult(), err2
}

// FetchScroll fetch additional data from an existing result
// using a scrollId.
func (c *ESClient) FetchScroll(scrollID string, ttl string) (Result, error) {
	jsonBody, err := json.Marshal(scrollObj{Scroll: ttl, ScrollID: scrollID})
	if err != nil {
		return NewEmptyResult(), err
	}
	resp, err := c.Do("POST", "/_search/scroll", jsonBody)
	if err != nil {
		return NewEmptyResult(), err
	}
	var srchResult Result
	err = json.Unmarshal(resp, &srchResult)
	if err != nil {
		return NewEmptyResult(), err
	}
	return srchResult, nil
}

// DocFilter specifies parameters of filtering operation
type DocFilter struct {
	AppType         string  `json:"appType"`
	Disabled        bool    `json:"disabled"`
	FromDate        string  `json:"fromDate"`
	ToDate          string  `json:"toDate"`
	IPAddress       string  `json:"ipAddress"`
	UserAgent       string  `json:"userAgent"`
	Action          string  `json:"action"`
	WithProbability float64 `json:"_withProbability"`
}

// SearchRecords searches records matching provided filter.
// Result fetching uses ElasticSearch scroll mechanism which requires
// providing TTL value to specify how long the result scroll should be
// available.
func (c *ESClient) SearchRecords(filter DocFilter, ttl string, chunkSize int) (Result, error) {
	encQuery, err := CreateClientSrchQuery(filter, chunkSize)
	if err == nil {
		return c.search(encQuery, ttl)
	}
	return NewEmptyResult(), err
}
