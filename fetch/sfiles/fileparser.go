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

package sfiles

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/czcorpus/klogproc/transform"
	"github.com/czcorpus/klogproc/transform/kontext"
)

type minorError struct {
	LineNumber int
	Message    string
}

func (m minorError) Error() string {
	return fmt.Sprintf("line %d: %s", m.LineNumber, m.Message)
}

func newMinorError(lineNumber int, message string) minorError {
	return minorError{LineNumber: lineNumber, Message: message}
}

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

// newParser creates a new instance of the Parser.
// localTimezone has format: "(-|+)[0-9]{2}:[0-9]{2}"
func newParser(path string, localTimezone string) *Parser {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	return &Parser{
		fr:            sc,
		localTimezone: localTimezone,
		fileName:      filepath.Base(f.Name()),
	}
}

// Parser parses a single file represented by fr Scanner.
// Because KonText does not log (at least currently) a timezone info,
// this information is also required to process the log properly.
type Parser struct {
	fr            *bufio.Scanner
	fileName      string
	localTimezone string
}

// parseLine parses a query log line - i.e. it expects
// that the line contains user interaction log
func (p *Parser) parseLine(s string, lineNum int) (*kontext.InputRecord, error) {
	jsonLine := parseRawLine(s)
	if jsonLine != "" {
		return kontext.ImportJSONLog([]byte(jsonLine), p.localTimezone)

	} else if tp := getLineType(s); tp == "QUERY" {
		return nil, fmt.Errorf("Failed to process QUERY entry: %s", s)

	} else {
		return nil, newMinorError(lineNum, fmt.Sprintf("ignored non-query entry"))
	}
}

// Parse runs the parsing process based on provided minimum accepted record
// time, record type (which is just passed to ElastiSearch) and a
// provided LogInterceptor).
func (p *Parser) Parse(fromTimestamp int64, recType string, proc transform.LogTransformer, outputs ...chan *kontext.OutputRecord) {
	for i := 0; p.fr.Scan(); i++ {
		rec, err := p.parseLine(p.fr.Text(), i)
		if err == nil {
			if rec.GetTime().Unix() >= fromTimestamp {
				outRec := proc.ProcItem(recType, rec)
				if outRec != nil {
					for _, output := range outputs {
						output <- outRec
					}
				}
			}

		} else {
			switch err.(type) {
			case minorError:
				log.Printf("INFO: file %s, %s", p.fileName, err)
			default:
				log.Print("ERROR: ", err)
			}

		}
	}
}
