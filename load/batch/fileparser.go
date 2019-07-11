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

package batch

import (
	"bufio"
	"log"
	"os"
	"path/filepath"

	"github.com/czcorpus/klogproc/transform"
)

// newParser creates a new instance of the Parser.
// localTimezone has format: "(-|+)[0-9]{2}:[0-9]{2}"
func newParser(path string, localTimezone string, appType string) *Parser {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	lineParser, err := newLineParser(appType)
	if err != nil {
		panic(err) // TODO
	}
	return &Parser{
		recType:       appType,
		fr:            sc,
		localTimezone: localTimezone,
		fileName:      filepath.Base(f.Name()),
		lineParser:    lineParser,
	}
}

// LineParser represents an object able to parse an individual
// line from a specific application log.
type LineParser interface {
	ParseLine(s string, lineNum int, localTimezone string) (transform.InputRecord, error)
}

// Parser parses a single file represented by fr Scanner.
// Because KonText does not log (at least currently) a timezone info,
// this information is also required to process the log properly.
type Parser struct {
	fr            *bufio.Scanner
	fileName      string
	localTimezone string
	lineParser    LineParser
	recType       string
}

// Parse runs the parsing process based on provided minimum accepted record
// time, record type (which is just passed to ElastiSearch) and a
// provided LogInterceptor).
func (p *Parser) Parse(fromTimestamp int64, proc LogItemProcessor, outputs ...chan transform.OutputRecord) {
	for i := 0; p.fr.Scan(); i++ {
		rec, err := p.lineParser.ParseLine(p.fr.Text(), i, p.localTimezone)
		if err == nil {
			if rec.GetTime().Unix() >= fromTimestamp {
				outRec := proc.ProcItem(p.recType, rec)
				if outRec != nil {
					for _, output := range outputs {
						output <- outRec
					}
				}
			}

		} else {
			switch err.(type) {
			case transform.MinorParsingError:
				log.Printf("INFO: file %s, %s", p.fileName, err)
			default:
				log.Print("ERROR: ", err)
			}

		}
	}
}
