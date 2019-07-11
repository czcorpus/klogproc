// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

package syd

import (
	"fmt"
	"strings"
)

// LineParser is a parser for reading KonText application logs
type LineParser struct {
}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	// %{TIMESTAMP_ISO8601:datetime}[\t]%{DATA:ipAddress}[\t]%{DATA:userId}[\t]%{DATA:keyReq}[\t]%{DATA:keyUsed}[\t]%{DATA:key}[\t]%{DATA:lTool}[\t]%{GREEDYDATA:runScript}"
	items := strings.Split(s, "\t")
	if len(items) >= 8 {
		return &InputRecord{
			Datetime:  items[0],
			IPAddress: items[1],
			UserID:    items[2],
			KeyReq:    items[3],
			KeyUsed:   items[4],
			Key:       items[5],
			Ltool:     items[6],
			RunScript: items[7],
		}, nil
	}
	return nil, fmt.Errorf("Invalid line format")
}
