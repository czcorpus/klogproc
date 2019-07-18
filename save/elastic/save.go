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
	"log"
	"sync"

	"github.com/czcorpus/klogproc/conversion"
)

const (
	es6DocType = "query"
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
		return fmt.Errorf("Failed to push log chunk: %s", err)
	}
	log.Printf("INFO: Inserted chunk of %d items to ElasticSearch\n", (len(data)-1)/2)
	return nil
}

// ----

// RunWriteConsumer reads incoming records from incomingData channel and writes them
// chunk by chunk. Once the channel is closed, the rest of items in buffer is writtten
// and the consumer finishes.
func RunWriteConsumer(appType string, conf *ConnectionConf, incomingData chan conversion.OutputRecord, waitGroup *sync.WaitGroup, failHandler ESImportFailHandler) {
	// Elasticsearch bulk writes
	defer waitGroup.Done()
	if conf.IsConfigured() {
		i := 0
		data := make([][]byte, conf.PushChunkSize*2+1)
		failed := make([][]byte, 0, 50)
		var esErr error
		for rec := range incomingData {
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

			if err == nil && err2 == nil {
				data[i] = jsonMetaES
				data[i+1] = jsonData
				i += 2

			} else {
				log.Print("ERROR: Failed to encode item ", rec.GetTime())
			}
			if i == conf.PushChunkSize*2 {
				data[i] = []byte("\n")
				esErr = BulkWriteRequest(data[:i+1], appType, conf)
				if esErr != nil {
					log.Print("ERROR: Failed to save a data chunk to ElasticSearch")
					failed = append(failed, data[:i+1]...)
				}
				i = 0
			}
		}
		if i > 0 {
			data[i] = []byte("\n")
			esErr = BulkWriteRequest(data[:i+1], appType, conf)
			if esErr != nil {
				log.Printf("ERROR: %s", esErr)
				failed = append(failed, data[:i+1]...)
			}
		}
		if failHandler != nil {
			failHandler.RescueFailedChunks(failed)
		}

	} else {
		for range incomingData {
		}
	}
}
