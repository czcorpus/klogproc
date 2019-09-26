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

package wag

import (
	"strings"

	"github.com/czcorpus/klogproc/load/accesslog"
)

var (
	pathPrefixes = []string{"slovo-v-kostce", "word-at-a-glance", "wag", "wdg", "wdglance"}
)

func getAction(path string) string {
	items := strings.Split(strings.Trim(path, "/"), "/")
	for _, prefix := range pathPrefixes {
		if items[0] == prefix && len(items) > 1 {
			return items[1]
		}
	}
	return items[0]
}

func isLoggable(action string) bool {
	return action == "search" || action == "word-forms" || action == "similar-freq-words"
}

// LineParser is a parser for reading KonText application logs
type LineParser struct {
	parser accesslog.LineParser
}

// ParseLine parses a HTTP access log format line
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	parsed, err := lp.parser.ParseLine(s, lineNum, localTimezone)
	if err != nil {
		return &InputRecord{isLoggable: false}, err
	}
	action := getAction(parsed.Path)
	if action == "" {
		return &InputRecord{isLoggable: false}, nil
	}

	ans := &InputRecord{
		isLoggable: isLoggable(action),
		Action:     action,
		Datetime:   parsed.Datetime,
		Request: Request{
			HTTPUserAgent:  parsed.UserAgent,
			HTTPRemoteAddr: parsed.IPAddress,
			RemoteAddr:     parsed.IPAddress, // TODO the same stuff as above?
		},
		ProcTime:  parsed.ProcTime,
		QueryType: parsed.URLArgs.Get("queryType"),
		Lang1:     parsed.URLArgs.Get("lang1"),
		Lang2:     parsed.URLArgs.Get("lang2"),
		Query1:    parsed.URLArgs.Get("q1"),
		Query2:    parsed.URLArgs.Get("q2"),
	}
	return ans, nil
}
