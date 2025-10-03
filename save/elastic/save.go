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
	"time"

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
	esclient := NewClient(esconf, appType)
	q := bytes.Join(data, []byte("\n"))
	_, err := esclient.DoBulkRequest("POST", "/_bulk", q)
	if err != nil {
		return fmt.Errorf("failed to push log chunk: %w", err)
	}
	log.Debug().Msgf("Inserted chunk of %d items to ElasticSearch", (len(data)-1)/2)
	return nil
}

// WriteBulkWithError is used for data where at least one error is expected.
// It splits data into two halfs and tries to insert them independently.
// Then it works recursively until chunks are inserted all small enough to
// stop. This allows for not dropping a whole chunk because of a single error
// (or few errors). The action itself is not recoverable so in case it fails
// from any reason, the items we wanted to write are definitely lost.
func WriteBulkWithError(data [][]byte, appType string, esconf *ConnectionConf) {
	if len(data) <= 10 {
		data = append(data, []byte("\n"))
		if err := BulkWriteRequest(data, appType, esconf); err != nil {
			log.Error().Err(err).Int("chunkSize", len(data)).Msg("failed to insert exploded chunk")

		} else {
			log.Info().Int("chunkSize", len(data)).Msg("successfully inserted exploded chunk ")
		}

	} else {
		if len(data)%2 == 1 { // => original chunk with newline at the end
			data = data[:len(data)-1]
		}
		split := len(data) / 2
		data1 := data[:split]
		time.Sleep(2 * time.Second)
		WriteBulkWithError(data1, appType, esconf)
		data2 := data[split:]
		time.Sleep(2 * time.Second)
		WriteBulkWithError(data2, appType, esconf)
	}
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
			appType = servicelog.MapAppTypeToIndex(appType)
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
					chunkPosition.Written = true
					if esErr != nil {
						dataCopy := make([][]byte, len(data[:i+1]))
						go func() {
							log.Warn().Err(esErr).Msg("due to inserting error, klogproc will try to insert smaller chunks")
							copy(dataCopy, data[:i+1])
							WriteBulkWithError(dataCopy, appType, conf)
						}()
					}
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
