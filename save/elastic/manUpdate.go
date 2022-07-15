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

	"github.com/rs/zerolog/log"
)

// DocUpdateFilter specifies parameters of docupdate operation
type DocUpdateFilter struct {
	AppType   string `json:"appType"`
	Disabled  bool   `json:"disabled"`
	FromDate  string `json:"fromDate"`
	ToDate    string `json:"toDate"`
	IPAddress string `json:"ipAddress"`
	UserAgent string `json:"userAgent"`
}

// DocUpdRecord is a general object providing update for an
// ElasticSearch record.
type DocUpdRecord map[string]interface{}

// DocUpdConf wraps both filters used to select
// records to be updated and also an update object
// used to merge with selected records.
type DocUpdConf struct {

	// Filters specifies which items should we look for.
	// Items in the list are taken as logical conjunction
	// (i.e. rule[0] && rule[1] && ... && rule[N])
	Filters []DocUpdateFilter `json:"filters"`

	// Update is a (sub)record we will use to write into
	// matching records.
	Update DocUpdRecord `json:"update"`

	// RemoveKey specifies a key we want to remove from
	// matching records. Functionally this value is
	// mutually exclusive with 'Update' (i.e. only one
	// of these operations can be defined at once).
	RemoveKey string `json:"removeKey"`

	// SearchChunkSize specifies how many items at once should
	// klogproc search and load for a specified update. For a slow
	// environment, keep the value reasonably small.
	SearchChunkSize int `json:"searchChunkSize"`
}

type docUpdObj struct {
	Doc DocUpdRecord `json:"doc"`
}

func (duo *docUpdObj) ToJSONQuery() ([]byte, error) {
	return json.Marshal(duo)
}

type docKeyRemoveObj struct {
	Script string `json:"script"`
}

func (dkr *docKeyRemoveObj) ToJSONQuery() ([]byte, error) {
	return json.Marshal(dkr)
}

type docBulkUpdateMetaObj struct {
	Update docBulkMetaRecord `json:"update"`
}

type docBulkMetaRecord struct {
	// "/"+c.index+"/"+item.Type+"/"+item.ID+"/_update", updQuery)
	Index string `json:"_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"`
}

// UpdResponse describes ElasticSearch response
// for the update call.
type UpdResponse struct {
	Index   string      `json:"_index"`
	Type    string      `json:"_type"`
	ID      string      `json:"_id"`
	Version int         `json:"_version"`
	Result  string      `json:"result"`
	Shards  interface{} `json:"_shards"` // we don't care much about this (yet)
}

func (c *ESClient) bulkUpdateUpdRecordScroll(index string, hits Hits, rawOp []byte) (int, error) {
	jsonLines := make([][]byte, len(hits.Hits)*2+1) // one for final 'new line'
	stopIdx := 0
	for _, item := range hits.Hits {
		jsonMeta, err := createDocBulkMetaRecord(index, item.Type, item.ID)
		if err != nil {
			log.Panic().Msgf("Failed to generate bulk update JSON (meta): %v", err)
		}
		jsonLines[stopIdx] = jsonMeta
		//log.Debug().Msgf("json meta: %s", string(jsonMeta))
		jsonLines[stopIdx+1] = rawOp
		//log.Debug().Msgf("json data: %s", string(jsonData))
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

func createDocBulkMetaRecord(index string, objType string, id string) ([]byte, error) {
	d := docBulkMetaRecord{Index: index, Type: objType, ID: id}
	obj := docBulkUpdateMetaObj{Update: d}
	return json.Marshal(obj)
}

func (c *ESClient) manualBulkRecordOp(index string, filters DocUpdateFilter, rawOp []byte, scrollTTL string, srchChunkSize int) (int, error) {
	totalUpdated := 0
	if !filters.Disabled {
		items, err := c.SearchRecords(filters, scrollTTL, srchChunkSize)
		if err != nil {
			return totalUpdated, err

		} else if items.Hits.Total == 0 {
			return 0, nil

		} else if len(items.Hits.Hits) > 0 {
			ans, bulkErr := c.bulkUpdateUpdRecordScroll(index, items.Hits, rawOp)
			totalUpdated += ans
			if bulkErr != nil {
				return totalUpdated, bulkErr
			}
		}
		if items.ScrollID != "" {
			for len(items.Hits.Hits) > 0 {
				items, err = c.FetchScroll(items.ScrollID, scrollTTL)
				if err != nil {
					return totalUpdated, err
				}
				if len(items.Hits.Hits) > 0 {
					ans, bulkErr := c.bulkUpdateUpdRecordScroll(index, items.Hits, rawOp)
					totalUpdated += ans
					if bulkErr != nil {
						return totalUpdated, err
					}
				}
			}
		}
	}
	return totalUpdated, nil
}

// ManualBulkRecordUpdate updates matching records with provided object
func (c *ESClient) ManualBulkRecordUpdate(index string, filters DocUpdateFilter, upd DocUpdRecord, scrollTTL string, srchChunkSize int) (int, error) {

	jsonData, err := createLogRecUpdQuery(upd)
	if err != nil {
		log.Fatal().Msgf("Failed to generate bulk update JSON (values): %s", err)
	}
	return c.manualBulkRecordOp(index, filters, jsonData, scrollTTL, srchChunkSize)
}

// ManualBulkRecordKeyRemove removes a specified key from matching records.
func (c *ESClient) ManualBulkRecordKeyRemove(index string, filters DocUpdateFilter, key string, scrollTTL string, srchChunkSize int) (int, error) {

	jsonData, err := createLogRecKeyRemoveQuery(key)
	if err != nil {
		log.Fatal().Msgf("Failed to generate bulk update JSON (values): %s", err)
	}
	return c.manualBulkRecordOp(index, filters, jsonData, scrollTTL, srchChunkSize)
}
