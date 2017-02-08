// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

package logs

import (
	"bufio"
	"encoding/json"
	// "fmt"
	"os"
	"regexp"
	"time"
)

type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
}

type LogRecord struct {
	UserID   int               `json:"user_id"`
	ProcTime float32           `json:"proc_time"`
	Date     string            `json:"date"`
	Action   string            `json:"action"`
	Request  Request           `json:"request"`
	Params   map[string]string `json:"params"`
}

func (rec *LogRecord) GetTime() time.Time {
	if t, err := time.Parse("2006-01-02 15:04:05", rec.Date); err == nil {
		return t
	}
	return time.Time{}
}

func parseRawLine(s string) string {
	reg := regexp.MustCompile("^.+\\sINFO:\\s+(\\{.+)$")
	srch := reg.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func NewParser(path string) *Parser {
	f, err := os.Open(path)
	if err == nil {
		sc := bufio.NewScanner(f)
		return &Parser{fr: sc}
	}
	panic(err)
}

type Parser struct {
	fr *bufio.Scanner
}

func (p *Parser) parseLine(s string) {
	jsonLine := parseRawLine(s)
	if jsonLine != "" {
		var record LogRecord
		err := json.Unmarshal([]byte(jsonLine), &record)
		if err == nil {
			//fmt.Println(record)
		}
	}
}

func (p *Parser) Parse(fromTimestamp int) {
	for p.fr.Scan() {
		p.parseLine(p.fr.Text())
	}
}
