// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 202 Institute of the Czech National Corpus,
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

package save

import (
	"fmt"
	"log"
	"sync"

	"github.com/czcorpus/klogproc/conversion"
)

// RunWriteConsumer reads from incomingData channel and stores the data
// to a configured InfluxDB measurement. For performance reasons, the actual
// database write is performed each time number of added items equals
// conf.PushChunkSize and also once the incomingData channel is closed.
func RunWriteConsumer(incomingESData <-chan conversion.OutputRecord, incomingInfluxData <-chan conversion.OutputRecord, waitGroup *sync.WaitGroup, confirmChan chan ConfirmMsg) {
	go func() {
		for range incomingInfluxData {
		}
		waitGroup.Done()
	}()

	defer waitGroup.Done()
	for item := range incomingESData {
		out, err := item.ToJSON()
		if err != nil {
			log.Print("ERROR: ", err)

		} else {
			fmt.Println(string(out))
		}
	}
}
