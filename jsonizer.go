// Copyright 2018 Tomas Machalek <tomas.machalek@gmail.com>
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/czcorpus/klogproc/fetch"
)

type NormalizedLogRecord struct {
	UserID         int                    `json:"user_id"`
	ProcTime       float32                `json:"proc_time"`
	Date           string                 `json:"date"`
	Action         string                 `json:"action"`
	Request        fetch.Request          `json:"request"`
	Params         map[string]interface{} `json:"params"`
	Settings       map[string]interface{} `json:"settings"`
	AlignedCorpora []string               `json:"alignedCorpora"`
}

type ExportProcessor struct {
}

func (ep *ExportProcessor) ProcItem(appType string, record *fetch.LogRecord) {
	if record.AgentIsLoggable() && record.Action == "first" {
		corpname, ok := (record.Params["corpname"]).(string)
		if ok && (strings.HasPrefix(corpname, "syn") || strings.HasPrefix(corpname, "omezeni/syn")) {
			outRec := NormalizedLogRecord{
				UserID:         record.UserID,
				ProcTime:       record.ProcTime,
				Date:           record.Date,
				Action:         record.Action,
				Request:        record.Request,
				Params:         record.Params,
				Settings:       record.Settings,
				AlignedCorpora: record.GetAlignedCorpora(),
			}
			s, err := json.Marshal(&outRec)
			if err != nil {
				log.Print("ERROR: ", err)

			} else {
				fmt.Println(string(s))
			}
		}
	}
}

func JsonizeLogs(conf *Conf) {
	proc := &ExportProcessor{}
	processFileLogs(conf, 0, proc)
}
