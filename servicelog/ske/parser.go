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

package ske

import (
	"strings"

	"klogproc/load/accesslog"
)

const (
	actionMark    = "run.cgi/"
	actionMarkLen = len("run.cgi/")
)

func getAction(path string) string {
	i := strings.Index(path, actionMark)
	if i > -1 {
		return path[i+actionMarkLen:]
	}
	return ""
}

// LineParser is a parser for reading SkE application logs
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
		isProcessable: true,
		Action:        action,
		Corpus:        parsed.URLArgs.Get("corpname"),
		Subcorpus:     parsed.URLArgs.Get("usesubcorp"),
		User:          parsed.Username,
		Datetime:      parsed.Datetime,
		Request: Request{
			HTTPUserAgent:  parsed.UserAgent,
			HTTPRemoteAddr: parsed.IPAddress,
			RemoteAddr:     parsed.IPAddress, // TODO the same stuff as above?
		},
		ProcTime: parsed.ProcTime,
	}
	return ans, nil
}
