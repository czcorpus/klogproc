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
)

type ESClientError struct {
	Message string
	Query   []byte
}

func (esc *ESClientError) Error() string {
	return esc.Message
}

func NewESClientError(message string, query []byte) *ESClientError {
	return &ESClientError{message, query}
}

// ESClient is a simple ElasticSearch client
type ESClient struct {
	server string
	index  string
}

// NewClient returns an instance of ESClient
func NewClient(server string, index string) *ESClient {
	c := ESClient{
		server: server,
		index:  index,
	}
	return &c
}

// UpdateSetAPIFlag sets a (new) attribute "isApi" to "true" for all
// the matching documents
func (c *ESClient) UpdateSetAPIFlag(conf APIFlagUpdateConf) ([]byte, error) {
	if !conf.Disabled {
		ans, err := CreateClientSrchQuery(conf.FromDate, conf.ToDate, conf.IPAddress, conf.UserAgent)
		if err == nil {
			fmt.Println("Q: ", string(ans))
			return c.Do("GET", "/"+c.index+"/_search", ans)
		}
		return make([]byte, 0), err
	}
	return make([]byte, 0), nil
}

// Do sends a general request to ElasticSearch server where
// 'query' is expected to be a JSON-encoded argument object
func (c *ESClient) Do(method string, path string, query []byte) ([]byte, error) {
	body := bytes.NewBuffer(query)
	client := http.Client{}
	req, err := http.NewRequest(method, c.server+path, body)
	if err != nil {
		return make([]byte, 0), err
	}
	resp, err := client.Do(req)
	if err != nil {
		return make([]byte, 0), err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return make([]byte, 0), NewESClientError(fmt.Sprintf("Invalid status code: %d", resp.StatusCode), query)
	}
	ans, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return ans, nil
}

func (c *ESClient) Search(query []byte) (Result, error) {
	resp, err := c.Do("GET", "/"+c.index+"/_search", query)
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

func (c *ESClient) SearchForAgents(conf APIFlagUpdateConf) (Result, error) {
	encQuery, err := CreateClientSrchQuery(conf.FromDate, conf.ToDate, conf.IPAddress, conf.UserAgent)
	if err == nil {
		return c.Search(encQuery)
	}
	return NewEmptyResult(), err
}
