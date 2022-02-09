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

package mapka

import (
	"strings"

	"klogproc/load/accesslog"
)

func getAction(path string) (string, *RequestParams) {
	var params *RequestParams
	if strings.HasPrefix(path, "/mapka") {
		elms := strings.Split(strings.Trim(path, "/"), "/")
		if len(elms) > 1 {
			if elms[1] == "text" {
				if len(elms) >= 4 {
					params = &RequestParams{
						CardType:   &elms[2],
						CardFolder: &elms[3],
					}
				}
				return elms[1], params

			} else if elms[1] == "overlay" {
				if len(elms) >= 3 {
					sub := strings.Split(elms[2], "+")
					params = &RequestParams{
						OverlayFile: &sub[len(sub)-1],
					}
				}
				return elms[1], params
			}
			return "", nil
		}
		return "index", params
	}
	return "", params
}

// LineParser is a parser for reading Mapka application logs
type LineParser struct {
	parser accesslog.LineParser
}

// ParseLine parses a HTTP access log format line
func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {
	parsed, err := lp.parser.ParseLine(s, lineNum)
	if err != nil {
		return &InputRecord{isProcessable: false}, err
	}

	action, params := getAction(parsed.Path)
	if action == "" {
		return &InputRecord{isProcessable: false}, nil
	}
	ans := &InputRecord{
		isProcessable: strings.HasPrefix(parsed.Path, "/mapka"),
		Action:        action,
		Path:          parsed.Path,
		Datetime:      parsed.Datetime,
		Request: &Request{
			HTTPUserAgent:  parsed.UserAgent,
			HTTPRemoteAddr: parsed.IPAddress,
			RemoteAddr:     parsed.IPAddress, // TODO the same stuff as above?
		},
		Params:   params,
		ProcTime: parsed.ProcTime,
	}
	return ans, nil
}
