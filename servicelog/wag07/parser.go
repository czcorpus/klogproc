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

package wag07

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

const (
	actionSearch           = "search"
	actionTranslate        = "translate"
	actionCompare          = "compare"
	actionWordForms        = "word-forms"
	actionSimilarFreqWords = "similar-freq-words"
	actionSetTheme         = "set-theme"
	actionTelemetry        = "telemetry"
	actionSetUILang        = "set-ui-lang"
	actionGetLemmas        = "get-lemmas"
	actionEmbedded         = "embedded"
)

// LineParser is a parser for reading KonText application logs
type LineParser struct {
}

func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {
	var record InputRecord
	err := json.Unmarshal([]byte(s), &record)
	if err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			// we ignore syntax errors as we expect some lines to contain garbage
			return &record, nil
		}
		return &record, err
	}
	if len(record.Timestamp) == 0 {
		log.Warn().Msg("missing time information in wag07 record, skipping")
		record.isProcessable = false

	} else {
		record.isProcessable = true
	}
	return &record, nil
}
