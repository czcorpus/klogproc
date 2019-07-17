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

package kwords

import (
	"fmt"
	"strings"

	"github.com/czcorpus/klogproc/conversion"
)

/*

 */

// LineParser is a parser for reading KWords application log
// which is basically a TAB separated list of items.
type LineParser struct {
}

// ParseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
// Format:
// {datetime_ISO8601}[TAB]{ipAddress}[TAB]{userId}[TAB]{numFiles}[TAB]
// {targetInputType}[TAB]{targetLength}[TAB]{corpus}[TAB]{refLength}[TAB]
// {pronouns}[TAB]{prep}[TAB]{con}[TAB]{num}[TAB]{caseInsensitive}
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	items := strings.Split(s, "\t")
	var err error

	if len(items) >= 13 {
		return &InputRecord{
			Datetime:        items[0],
			IPAddress:       items[1],
			UserID:          items[2],
			NumFiles:        items[3],
			TargetInputType: items[4],
			TargetLength:    items[5],
			Corpus:          items[6],
			RefLength:       items[7],
			Pronouns:        items[8],
			Prep:            items[9],
			Con:             items[10],
			Num:             items[11],
			CaseInsensitive: items[12],
		}, err
	}
	return nil, conversion.NewMinorParsingError(
		lineNum,
		fmt.Sprintf("Invalid line format for KWords. Expecting 13 tab-separated items, found %d", len(items)),
	)
}
