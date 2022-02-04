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

	"klogproc/conversion"
)

// RunWriteConsumer runs a dummy (null) write consumer
func RunWriteConsumer(incomingData <-chan *conversion.BoundOutputRecord) <-chan ConfirmMsg {
	confirmChan := make(chan ConfirmMsg)
	defer func() {
		close(confirmChan)
	}()
	go func() {
		for item := range incomingData {
			out, err := item.ToJSON()
			if err != nil {
				log.Print("ERROR: ", err)

			} else {
				fmt.Println(string(out))
			}
		}
	}()
	return confirmChan
}
