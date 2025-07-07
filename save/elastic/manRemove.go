// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	docRmDefaultSearchChunkSize = 10000
)

// DocRemConf wraps filters used to reove records
type DocRemConf struct {
	AppType string `json:"appType"`

	// Filters specifies which items should we look for.
	// Items in the list are taken as logical conjunction
	// (i.e. rule[0] && rule[1] && ... && rule[N])
	Filters []DocFilter `json:"filters"`

	// SearchChunkSize specifies how many items at once should
	// klogproc search and load for a specified cleaning. For a slow
	// environment, keep the value reasonably small.
	SearchChunkSize int `json:"searchChunkSize"`
}

func (df *DocRemConf) Overview() string {
	var buff strings.Builder
	for i, filter := range df.Filters {
		buff.WriteString(fmt.Sprintf("%d)\n%s\n", i+1, filter.Overview()))
	}
	return buff.String()
}

func (df *DocRemConf) Validate() error {
	if df.SearchChunkSize == 0 {
		log.Warn().
			Int("default", docRmDefaultSearchChunkSize).
			Msg("recordRemove.searchChunkSize not specified, using default")
		df.SearchChunkSize = docRmDefaultSearchChunkSize
	}
	if df.AppType == "" {
		return fmt.Errorf("missing appType")
	}
	for index, filter := range df.Filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("filter #%d: %w", index, err)
		}
	}
	return nil
}

type docBulkRemoveMetaObj struct {
	Delete docBulkMetaRecord `json:"delete"`
}

func createDocBulkRemoveMetaRecord(index string, objType string, id string) ([]byte, error) {
	d := docBulkMetaRecord{Index: index, Type: objType, ID: id}
	obj := docBulkRemoveMetaObj{Delete: d}
	return json.Marshal(obj)
}

func (c *ESClient) bulkRemoveRecordScroll(index string, hits Hits) (int, error) {
	jsonLines := make([][]byte, len(hits.Hits)+1) // one for final 'new line'
	stopIdx := 0
	for _, item := range hits.Hits {
		jsonMeta, err := createDocBulkRemoveMetaRecord(index, item.Type, item.ID)
		if err != nil {
			log.Panic().Msgf("Failed to generate bulk remove JSON (meta): %v", err)
		}
		jsonLines[stopIdx] = jsonMeta
		// log.Debug().Msgf("json meta: %s", string(jsonMeta))
		stopIdx += 1
	}
	jsonLines[stopIdx] = make([]byte, 0)
	stopIdx++
	_, err := c.Do("POST", "/_bulk", bytes.Join(jsonLines[:stopIdx], []byte("\n")))
	if err != nil {
		return 0, err
	}
	return stopIdx, nil
}

// ManualBulkRecordRemove removes matching records
func (c *ESClient) ManualBulkRecordRemove(
	index string,
	appType string,
	filters DocFilter,
	scrollTTL string,
	srchChunkSize int,
	dryRun bool,
) (int, error) {
	totalRemoved := 0
	if !filters.Disabled {
		items, err := c.SearchRecords(appType, filters, scrollTTL, srchChunkSize)
		if filters.WithProbability > 0 {
			items.Hits = items.Hits.Sampled(filters.WithProbability)
		}
		if err != nil {
			return totalRemoved, err

		} else if items.Hits.Total == 0 {
			return 0, nil

		} else if len(items.Hits.Hits) > 0 {
			if dryRun {
				for _, hit := range items.Hits.Hits {
					log.Info().Any("item", hit).Msg("remove candidate")
				}
				totalRemoved += len(items.Hits.Hits)
			} else {
				ans, bulkErr := c.bulkRemoveRecordScroll(index, items.Hits)
				totalRemoved += ans
				if bulkErr != nil {
					return totalRemoved, bulkErr
				}
			}
		}
		if items.ScrollID != "" {
			for len(items.Hits.Hits) > 0 {
				items, err = c.FetchScroll(items.ScrollID, scrollTTL)
				if filters.WithProbability > 0 {
					items.Hits = items.Hits.Sampled(filters.WithProbability)
				}
				if err != nil {
					return totalRemoved, err
				}
				if len(items.Hits.Hits) > 0 {
					if dryRun {
						for _, hit := range items.Hits.Hits {
							log.Info().Any("item", hit).Msg("remove candidate")

						}
						totalRemoved += len(items.Hits.Hits)
					} else {
						ans, bulkErr := c.bulkRemoveRecordScroll(index, items.Hits)
						totalRemoved += ans
						if bulkErr != nil {
							return totalRemoved, bulkErr
						}
					}
				}
			}
		}
	}
	return totalRemoved, nil
}
