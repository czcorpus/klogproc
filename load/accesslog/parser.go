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

package accesslog

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/czcorpus/klogproc/conversion"
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

func getProcTime(procTimeExpr string) (float32, error) {
	srch := strings.Index(procTimeExpr, "rt=")
	if srch == 0 {
		pts := strings.Trim(procTimeExpr[3:], "\"")
		pt, err := strconv.ParseFloat(pts, 32)
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
		if len(item) == 0 {
			continue
		}
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

// ParsedAccessLog represents a general processing of an access log line
// without any dependency on a concrete Input implementation.
type ParsedAccessLog struct {
	IPAddress   string
	Username    string
	Datetime    string
	HTTPMethod  string
	HTTPVersion string
	Path        string
	URLArgs     url.Values
	Referrer    string
	UserAgent   string
	ProcTime    float32
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
func (lp *LineParser) ParseLine(s string, lineNum int, localTimezone string) (*ParsedAccessLog, error) {
	ans := &ParsedAccessLog{}
	var err error
	tokens := lp.tokenize(s)

	ans.IPAddress = tokens[0]
	ans.Username = tokens[2]
	ans.Datetime = tokens[3]
	urlBlock := strings.Split(tokens[4], " ")
	ans.HTTPMethod = urlBlock[0]
	ans.HTTPVersion = urlBlock[2]
	parsedURL, err := url.Parse(urlBlock[1])
	if err != nil {
		return nil, conversion.NewLineParsingError(lineNum, err.Error())
	}
	ans.Path = parsedURL.Path
	ans.URLArgs, err = url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		return nil, conversion.NewLineParsingError(lineNum, err.Error())
	}
	ans.Referrer = tokens[7]
	ans.UserAgent = tokens[8]
	ans.ProcTime, err = getProcTime(tokens[9])
	return ans, err
}
