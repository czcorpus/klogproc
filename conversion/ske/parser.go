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
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/czcorpus/klogproc/conversion"
)

const (
	actionMark    = "run.cgi/"
	actionMarkLen = len("run.cgi/")
)

func testOpenQuot(c byte) byte {
	switch c {
	case '"':
		return '"'
	case '[':
		return ']'
	default:
		return 0
	}
}

func isCloseQuot(c byte) bool {
	return c == '"' || c == ']'
}

func getAction(path string) string {
	i := strings.Index(path, actionMark)
	if i > -1 {
		return path[i+actionMarkLen:]
	}
	return ""
}

func getProcTime(procTimeExpr string) (float32, error) {
	srch := strings.Index(procTimeExpr, "rt=")
	if srch == 0 {
		pt, err := strconv.ParseFloat(procTimeExpr[3:], 32)
		if err != nil {
			return -1, fmt.Errorf("Failed to parse proc. time %s: %s", procTimeExpr, err)
		}
		return float32(pt), nil
	}
	return -1, fmt.Errorf("Failed to parse proc. time %s", procTimeExpr)
}

// LineParser is a parser for reading KonText application logs
type LineParser struct{}

func (lp *LineParser) tokenize(s string) []string {
	items := make([]string, 10)
	currQuoted := make([]string, 0, 30)
	var currQuotChar byte
	parsedPos := 0
	for _, item := range strings.Split(s, " ") {
		if currQuotChar == 0 {
			closeChar := testOpenQuot(item[0])
			if closeChar != 0 && item[len(item)-1] != closeChar {
				currQuoted = append(currQuoted, item[1:])
				currQuotChar = item[0]

			} else if closeChar != 0 && item[len(item)-1] == closeChar {
				items[parsedPos] = item[1 : len(item)-1]
				parsedPos++

			} else if closeChar == 0 {
				items[parsedPos] = item
				parsedPos++
			}

		} else {
			if isCloseQuot(item[len(item)-1]) {
				currQuoted = append(currQuoted, item[:len(item)-1])
				items[parsedPos] = strings.Join(currQuoted, " ")
				currQuotChar = 0
				parsedPos++
				currQuoted = make([]string, 0, 30)

			} else if !isCloseQuot(item[0]) && !isCloseQuot(item[len(item)-1]) {
				currQuoted = append(currQuoted, item)
			}
		}
	}
	return items
}

// ParseLine parses a HTTP access log format line
// data example:
//   0) 195.113.53.123
//   1) -
//   2) johndoe
//   3) [16/Sep/2019:08:24:05 +0200]
//   4) "GET /ske/css/images/ui-bg_highlight-hard_100_f2f5f7_1x100.png HTTP/2.0"
//   5) 200
//   6) 332
//   7) "https://www.korpus.cz/ske/css/jquery-ui.min.css"
//   8) "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/76.0.3809.100 Chrome/76.0.3809.100 Safari/537.36"
//   9) rt=0.012
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*InputRecord, error) {
	tokens := lp.tokenize(s)
	urlPart := strings.Split(tokens[4], " ")[1]
	parsedURL, err := url.Parse(urlPart)
	if err != nil {
		return nil, conversion.NewLineParsingError(lineNum, err.Error())
	}
	args, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return nil, conversion.NewLineParsingError(lineNum, err.Error())
	}
	action := getAction(parsedURL.Path)
	if action == "" {
		return &InputRecord{isLoggable: false}, nil
	}
	procTime, err := getProcTime(tokens[9])
	if err != nil && tokens[9] != "" {
		return &InputRecord{isLoggable: false}, err
	}

	ans := &InputRecord{
		isLoggable: true,
		Action:     action,
		Corpus:     args.Get("corpname"),
		Subcorpus:  args.Get("usesubcorp"),
		User:       tokens[2],
		Datetime:   tokens[3],
		Request: Request{
			HTTPUserAgent:  tokens[8],
			HTTPRemoteAddr: tokens[0],
			RemoteAddr:     tokens[0],
		},
		ProcTime: procTime,
	}
	return ans, nil
}
