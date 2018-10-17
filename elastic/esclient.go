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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SearchConf defines a configuration
// required to work with ES client.
type SearchConf struct {
	Server          string `json:"server"`
	Index           string `json:"index"`
	SearchChunkSize int    `json:"searchChunkSize"`
	PushChunkSize   int    `json:"pushChunkSize"`
	ScrollTTL       string `json:"scrollTtl"`
	ReqTimeoutSecs  int    `json:"reqTimeoutSecs"`
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
	Message string
	Query   []byte
	ESError ErrorResultObj
}

func (esc *ESClientError) Error() string {
	return fmt.Sprintf("%s: %s", esc.Message, esc.ESError)
}

func newESClientError(message string, response []byte, query []byte) *ESClientError {
	var errResult ErrorResultObj
	json.Unmarshal(response, &errResult)
	return &ESClientError{message, query, errResult}
}

// ESClient is a simple ElasticSearch client
type ESClient struct {
	server         string
	index          string
	srchChunkSize  int
	reqTimeoutSecs int
}

// NewClient returns an instance of ESClient
func NewClient(conf *SearchConf) *ESClient {
	c := ESClient{
		server:         conf.Server,
		index:          conf.Index,
		srchChunkSize:  conf.SearchChunkSize,
		reqTimeoutSecs: conf.ReqTimeoutSecs,
	}
	return &c
}

func (c ESClient) String() string {
	return fmt.Sprintf("ElasticSearchClient[server: %s, index; %s]", c.server, c.index)
}

// Do sends a general request to ElasticSearch server where
// 'query' is expected to be a JSON-encoded argument object
func (c *ESClient) Do(method string, path string, query []byte) ([]byte, error) {
	body := bytes.NewBuffer(query)
	client := http.Client{Timeout: time.Second * time.Duration(c.reqTimeoutSecs)}
	req, err := http.NewRequest(method, c.server+path, body)
	if err != nil {
		return make([]byte, 0), err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return make([]byte, 0), err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		errBody, _ := ioutil.ReadAll(resp.Body)
		return errBody, newESClientError(fmt.Sprintf("Request %s failed with code %d", path, resp.StatusCode), errBody, query)
	}
	ans, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return ans, nil
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

// SearchRecords searches records matching provided filter.
// Result fetching uses ElasticSearch scroll mechanism which requires
// providing TTL value to specify how long the result scroll should be
// available.
func (c *ESClient) SearchRecords(filter DocUpdateFilter, ttl string) (Result, error) {
	encQuery, err := CreateClientSrchQuery(filter.FromDate, filter.ToDate, filter.IPAddress, filter.UserAgent,
		c.srchChunkSize)
	if err == nil {
		return c.search(encQuery, ttl)
	}
	return NewEmptyResult(), err
}
