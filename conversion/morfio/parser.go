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

package morfio

import (
	"fmt"
	"strings"

	"github.com/czcorpus/klogproc/conversion"
)

// LineParser is a parser for reading Morfio application log
// which is basically a TAB separated list of items.
type LineParser struct {
}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
// Format:
// {datetime_ISO8601}[TAB]{ipAddress}[TAB]{userId}[TAB]{keyReq}[TAB]{keyUsed}[TAB]
// {key}[TAB]{runScript}[TAB]{corpus}[TAB]{minFreq}[TAB]{inputAttr}[TAB]{outputAttr}[TAB]
// {caseInsensitive}[TAB].*
func (lp *LineParser) ParseLine(s string, lineNum int) (*InputRecord, error) {

	items := strings.Split(s, "\t")
	var err error

	if len(items) >= 12 {
		return &InputRecord{
			Datetime:        items[0],
			IPAddress:       items[1],
			UserID:          items[2],
			KeyReq:          items[3],
			KeyUsed:         items[4],
			Key:             items[5],
			RunScript:       items[6],
			Corpus:          items[7],
			MinFreq:         items[8],
			InputAttr:       items[9],
			OutputAttr:      items[10],
			CaseInsensitive: items[11],
		}, err
	}
	return nil, conversion.NewLineParsingError(
		lineNum, fmt.Sprintf("Invalid line format for Morfio. Expecting 12 tab-separated items, found %d", len(items)))
}
