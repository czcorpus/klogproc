// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package influx

import (
	"log"
	"sync"

	"github.com/czcorpus/klogproc/conversion"
)

// RunWriteConsumer reads from incomingData channel and stores the data
// to a configured InfluxDB measurement. For performance reasons, the actual
// database write is performed each time number of added items equals
// conf.PushChunkSize and also once the incomingData channel is closed.
func RunWriteConsumer(conf *ConnectionConf, incomingData <-chan conversion.OutputRecord, waitGroup *sync.WaitGroup) {
	// InfluxDB batch writes
	if waitGroup != nil {
		defer waitGroup.Done()
	}
	if conf.IsConfigured() {
		var err error
		client, err := NewRecordWriter(conf)
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
		for rec := range incomingData {
			err = client.AddRecord(rec)
			if err != nil {
				log.Print("ERROR: ", err) // TODO
			}
		}
		err = client.Finish()
		if err != nil {
			log.Printf("ERROR: %s", err)
		}

	} else {
		for range incomingData {
		}
	}
}
