// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

// newParser creates a new instance of the Parser.
// tzShift can be used to correct an incorrectly stored datetime
func newParser(path string, tzShift int, appType string, version string, appErrRegister conversion.AppErrorRegister) *Parser {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	lineParser, err := NewLineParser(appType, version, appErrRegister)
	if err != nil {
		panic(err) // TODO
	}
	return &Parser{
		recType:    appType,
		fr:         sc,
		tzShift:    tzShift,
		fileName:   filepath.Base(f.Name()),
		lineParser: lineParser,
	}
}

// LineParser represents an object able to parse an individual
// line from a specific application log.
type LineParser interface {
	ParseLine(s string, lineNum int) (conversion.InputRecord, error)
}

// Parser parses a single file represented by fr Scanner.
// Because KonText does not log (at least currently) a timezone info,
// this information is also required to process the log properly.
type Parser struct {
	fr         *bufio.Scanner
	fileName   string
	tzShift    int
	lineParser LineParser
	recType    string
}

// Parse runs the parsing process based on provided minimum accepted record
// time, record type (which is just passed to ElastiSearch) and a
// provided LogInterceptor).
func (p *Parser) Parse(fromTimestamp int64, proc LogItemProcessor, fromTime, toTime *time.Time, outputs ...chan conversion.OutputRecord) {
	for i := 0; p.fr.Scan(); i++ {
		rec, err := p.lineParser.ParseLine(p.fr.Text(), i)
		if err == nil {
			recTime := rec.GetTime()
			if fromTime != nil && recTime.Before(*fromTime) {
				continue
			}
			if toTime != nil && recTime.After(*toTime) {
				continue
			}
			if recTime.Unix() >= fromTimestamp {
				outRec := proc.ProcItem(rec, p.tzShift)
				if outRec != nil {
					for _, output := range outputs {
						output <- outRec
					}
				}
			}

		} else {
			switch tErr := err.(type) {
			case conversion.LineParsingError:
				log.Printf("INFO: file %s, %s", p.fileName, tErr)
			default:
				log.Print("ERROR: ", tErr)
			}

		}
	}
}
