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

// fileparser contains functions used to parse KonText's file-stored
// application logs. In the recent KonText and Klogproc versions, this
// is rather a fallback solution as the logs are stored and read from
// a Redis queue (see redis.go).

package fetch

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

// LogInterceptor defines an object which is able to
// process individual LogRecord instances
type LogInterceptor interface {
	ProcItem(appType string, record *LogRecord)
}

// ------------------------------------------------------------

func parseRawLine(s string) string {
	reg := regexp.MustCompile("^.+\\sINFO:\\s+(\\{.+)$")
	srch := reg.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func getLineType(s string) string {
	reg := regexp.MustCompile("^.+\\s([A-Z]+):\\s+.+$")
	srch := reg.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func importDatetimeString(dateStr string, localTimezone string) (string, error) {
	rg := regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2})(\\s|T)([012]\\d:[0-5]\\d:[0-5]\\d\\.\\d+)")
	srch := rg.FindStringSubmatch(dateStr)
	if len(srch) > 0 {
		return fmt.Sprintf("%sT%s%s", srch[1], srch[3], localTimezone), nil
	}
	return "", fmt.Errorf("Failed to import datetime \"%s\"", dateStr)
}

// NewParser creates a new instance of the Parser.
// localTimezone has format: "(-|+)[0-9]{2}:[0-9]{2}"
func NewParser(path string, localTimezone string) *Parser {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	return &Parser{fr: sc, localTimezone: localTimezone}
}

// Parser parses a single file represented by fr Scanner.
// Because KonText does not log (at least currently) a timezone info,
// this information is also required to process the log properly.
type Parser struct {
	fr            *bufio.Scanner
	localTimezone string
}

// parseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
func (p *Parser) parseLine(s string) (*LogRecord, error) {
	jsonLine := parseRawLine(s)
	if jsonLine != "" {
		return ImportJSONLog([]byte(jsonLine), p.localTimezone)

	} else if tp := getLineType(s); tp == "QUERY" {
		return nil, fmt.Errorf("Failed to process QUERY entry: %s", s)

	} else {
		return nil, fmt.Errorf("Unrecognized entry: %s", s)
	}
}

// Parse runs the parsing process based on provided minimum accepted record
// time, record type (which is just passed to ElastiSearch) and a
// provided LogInterceptor).
func (p *Parser) Parse(fromTimestamp int64, recType string, proc LogInterceptor) {
	for p.fr.Scan() {
		rec, err := p.parseLine(p.fr.Text())
		if err == nil {
			if rec.GetTime().Unix() >= fromTimestamp {
				proc.ProcItem(recType, rec)
			}

		} else {
			log.Print("ERROR: ", err)
		}
	}
}
