// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package wsserver

import (
	"encoding/json"
	"regexp"

	"github.com/rs/zerolog/log"
)

var (
	recMatch = regexp.MustCompile("^.+ QUERY: (\\{.+)+")
)

type LineParser struct {
}

func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {
	srch := recMatch.FindStringSubmatch(s)
	ans := &InputRecord{isProcessable: false}
	if len(srch) > 0 {
		err := json.Unmarshal([]byte(srch[1]), ans)
		if err != nil {
			log.Error().Err(err).Msg("")

		} else {
			ans.isProcessable = true
			if ans.Datetime[len(ans.Datetime)-1] == 'Z' { // UTC time
				ans.Datetime = ans.Datetime[:len(ans.Datetime)-1]
			}
		}

	}
	return ans, nil
}
