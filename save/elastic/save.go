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
	"fmt"

	"klogproc/save"
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
)

const (
	es6DocType = "_doc"
)

// ESImportFailHandler represents an object able to handle (valid)
// log items we failed to insert to ElasticSearch (typically due
// to inavailability)
type ESImportFailHandler interface {
	RescueFailedChunks(chunk [][]byte) error
}

// ----

func BulkWriteRequest(data [][]byte, appType string, esconf *ConnectionConf) error {
	var esclient *ESClient
	if esconf.MajorVersion < 6 {
		esclient = NewClient(esconf)

	} else {
		esclient = NewClient6(esconf, appType)
	}
	q := bytes.Join(data, []byte("\n"))
	_, err := esclient.Do("POST", "/_bulk", q)
	if err != nil {
		return fmt.Errorf("failed to push log chunk: %w", err)
	}
	log.Debug().Msgf("Inserted chunk of %d items to ElasticSearch", (len(data)-1)/2)
	return nil
}

// ----

// RunWriteConsumer reads incoming records from incomingData channel and writes them
// chunk by chunk. Once the channel is closed, the rest of items in buffer is writtten
// and the consumer finishes.
func RunWriteConsumer(appType string, conf *ConnectionConf, incomingData <-chan *servicelog.BoundOutputRecord) <-chan save.ConfirmMsg {
	// Elasticsearch bulk writes
	confirmChan := make(chan save.ConfirmMsg)
	go func() {
		if conf.IsConfigured() {
			i := 0
			data := make([][]byte, conf.PushChunkSize*2+1)
			var chunkPosition *servicelog.LogRange
			var esErr error
			var rec *servicelog.BoundOutputRecord
			for rec = range incomingData {
				if chunkPosition == nil {
					chunkPosition = &rec.FilePos
				}
				chunkPosition.SeekEnd = rec.FilePos.SeekEnd
				jsonData, err := rec.ToJSON()
				recType := es6DocType
				index := fmt.Sprintf("%s_%s", conf.Index, appType)
				if conf.MajorVersion < 6 {
					recType = rec.GetType()
					index = conf.Index
				}
				jsonMeta := CNKRecordMeta{
					ID:    rec.GetID(),
					Type:  recType,
					Index: index,
				}
				jsonMetaES, err2 := (&ESCNKRecordMeta{Index: jsonMeta}).ToJSON()

				if err != nil {
					log.Error().Err(err).Msgf("Failed to encode item %s", rec.GetID())

				} else if err2 != nil {
					log.Error().Err(err2).Msgf("Failed to encode a 'meta' record for item %s", rec.GetID())

				} else {
					data[i] = jsonMetaES
					data[i+1] = jsonData
					i += 2
				}
				if i == conf.PushChunkSize*2 {
					data[i] = []byte("\n")
					esErr = BulkWriteRequest(data[:i+1], appType, conf)
					chunkPosition.Written = esErr == nil
					confirmMsg := save.ConfirmMsg{
						FilePath: rec.FilePath,
						Position: *chunkPosition,
						Error:    esErr,
					}
					confirmChan <- confirmMsg
					i = 0
				}
			}
			if i > 0 {
				data[i] = []byte("\n")
				esErr = BulkWriteRequest(data[:i+1], appType, conf)
				chunkPosition.Written = esErr == nil
				confirmMsg := save.ConfirmMsg{
					FilePath: rec.FilePath,
					Position: *chunkPosition,
					Error:    esErr,
				}
				confirmChan <- confirmMsg
			}
			close(confirmChan)

		} else {
			for range incomingData {
			}
			close(confirmChan)
		}
	}()
	return confirmChan
}
