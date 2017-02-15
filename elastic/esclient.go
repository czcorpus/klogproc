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
	"log"
	"net/http"
)

type ElasticSearchConf struct {
	ElasticServer          string `json:"elasticServer"`
	ElasticIndex           string `json:"elasticIndex"`
	ElasticSearchChunkSize int    `json:"elasticSearchChunkSize"`
	ElasticPushChunkSize   int    `json:"elasticPushChunkSize"`
	ElasticScrollTTL       string `json:"elasticScrollTtl"`
}

type ESClientError struct {
	Message string
	Query   []byte
	ESError ErrorResultObj
}

func (esc *ESClientError) Error() string {
	return fmt.Sprintf("%s, reason: %s", esc.Message, esc.ESError.Error.Reason)
}

func NewESClientError(message string, response []byte, query []byte) *ESClientError {
	var errResult ErrorResultObj
	json.Unmarshal(response, &errResult)
	return &ESClientError{message, query, errResult}
}

// ESClient is a simple ElasticSearch client
type ESClient struct {
	server        string
	index         string
	srchChunkSize int
}

// NewClient returns an instance of ESClient
func NewClient(server string, index string, srchChunkSize int) *ESClient {
	c := ESClient{
		server:        server,
		index:         index,
		srchChunkSize: srchChunkSize,
	}
	return &c
}

func (c *ESClient) Stringer() string {
	return fmt.Sprintf("ElasticSearchClient[server: %s, index; %s]", c.server, c.index)
}

func (c *ESClient) BulkUpdateSetAPIFlag(index string, conf APIFlagUpdateConf, scrollTTL string) (int, error) {
	totalUpdated := 0
	if !conf.Disabled {
		items, err := c.SearchForAgents(conf, scrollTTL)
		if err != nil {
			return totalUpdated, err
		}
		ans, bulkErr := c.bulkUpdateSetAPIFlagScroll(index, items.Hits)
		totalUpdated += ans
		if bulkErr != nil {
			return totalUpdated, bulkErr
		}
		if items.ScrollID != "" {
			for len(items.Hits.Hits) > 0 {
				items, err = c.FetchScroll(items.ScrollID, scrollTTL)
				if err != nil {
					return totalUpdated, err
				}
				if len(items.Hits.Hits) > 0 {
					ans, bulkErr = c.bulkUpdateSetAPIFlagScroll(index, items.Hits)
					totalUpdated += ans
					if err != nil {
						return totalUpdated, err
					}
				}
			}
		}
	}
	return totalUpdated, nil
}

func (c *ESClient) bulkUpdateSetAPIFlagScroll(index string, hits Hits) (int, error) {
	jsonLines := make([][]byte, len(hits.Hits)*2+1) // one for final 'new line'
	stopIdx := 0
	for _, item := range hits.Hits {
		jsonMeta, err := CreateDocBulkMetaRecord(index, item.Type, item.ID)
		if err != nil {
			log.Panicf("Failed to generate bulk update JSON (meta): %v", err)
		}
		jsonData, err := CreateClientAPIFlagUpdQuery()
		if err != nil {
			log.Panicf("Failed to generate bulk update JSON (values): %v", err)
		}
		jsonLines[stopIdx] = jsonMeta
		jsonLines[stopIdx+1] = jsonData
		stopIdx += 2
	}
	jsonLines[stopIdx] = make([]byte, 0)
	stopIdx++
	_, err := c.Do("POST", "/_bulk", bytes.Join(jsonLines[:stopIdx], []byte("\n")))
	if err != nil {
		return 0, err
	}
	return ((stopIdx - 1) / 2), nil
}

// UpdateSetAPIFlag sets a (new) attribute "isApi" to "true" for all
// the matching documents
func (c *ESClient) UpdateSetAPIFlag(conf APIFlagUpdateConf) ([]UpdResponse, error) {
	if !conf.Disabled {
		items, err := c.SearchForAgents(conf, "")
		if err != nil {
			return make([]UpdResponse, 0), err
		}
		responses := make([]UpdResponse, len(items.Hits.Hits))
		for i, item := range items.Hits.Hits {
			updQuery, err := CreateClientAPIFlagUpdQuery()
			if err != nil {
				return responses[:i], err
			}
			ans, err2 := c.Do("POST", "/"+c.index+"/"+item.Type+"/"+item.ID+"/_update", updQuery)
			if err2 != nil {
				return responses[:i], err2
			}
			var respObj UpdResponse
			if err3 := json.Unmarshal(ans, &respObj); err != nil {
				return responses[:i], err3
			}
			responses[i] = respObj
		}
		return responses, err
	}
	return make([]UpdResponse, 0), nil
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
		errBody, _ := ioutil.ReadAll(resp.Body)
		return errBody, NewESClientError(fmt.Sprintf("Invalid status code: %d", resp.StatusCode), query, errBody)
	}
	ans, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return ans, nil
}

func (c *ESClient) Search(query []byte, scroll string) (Result, error) {
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

func (c *ESClient) FetchScroll(scrollId string, ttl string) (Result, error) {
	jsonBody, err := json.Marshal(scrollObj{Scroll: ttl, ScrollID: scrollId})
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

func (c *ESClient) SearchForAgents(conf APIFlagUpdateConf, ttl string) (Result, error) {
	encQuery, err := CreateClientSrchQuery(conf.FromDate, conf.ToDate, conf.IPAddress, conf.UserAgent,
		c.srchChunkSize)
	if err == nil {
		return c.Search(encQuery, ttl)
	}
	return NewEmptyResult(), err
}
