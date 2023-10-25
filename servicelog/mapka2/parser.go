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

package mapka2

import (
	"regexp"
	"strings"

	"klogproc/load/accesslog"
)

var (
	suffPatt = regexp.MustCompile(`^.+\.([a-zA-Z0-9]+)$`)
)

func pathIsStaticFile(path string) bool {
	return suffPatt.MatchString(path)
}

func getAction(fullPath string) string {
	path := strings.Split(fullPath, "?")[0]
	if strings.HasPrefix(path, "/mapka") && !pathIsStaticFile(path) {
		elms := strings.Split(strings.Trim(path, "/"), "/")
		if len(elms) > 1 {
			return elms[1]
		}
		return "index"
	}
	return ""
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

	action := getAction(parsed.Path)
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
		ProcTime: parsed.ProcTime,
	}
	return ans, nil
}
