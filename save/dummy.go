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
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
)

// RunWriteConsumer runs a dummy (null) write consumer
func RunWriteConsumer(incomingData <-chan *servicelog.BoundOutputRecord, printOut bool) <-chan ConfirmMsg {
	confirmChan := make(chan ConfirmMsg)
	go func() {
		var chunkPosition *servicelog.LogRange
		for item := range incomingData {
			var jsonError error
			if chunkPosition == nil {
				chunkPosition = &item.FilePos
			}
			chunkPosition.SeekEnd = item.FilePos.SeekEnd
			out, jsonError := item.ToJSON()
			if jsonError != nil {
				log.Error().Err(jsonError).Msg("")

			} else {
				if printOut {
					fmt.Println(string(out))
				}
			}
			chunkPosition.Written = true
			confirmMsg := ConfirmMsg{
				FilePath: item.FilePath,
				Position: *chunkPosition,
				Error:    jsonError,
			}
			confirmChan <- confirmMsg
		}
		defer func() {
			close(confirmChan)
		}()
	}()
	return confirmChan
}
