// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

package sfiles

import (
	"fmt"

	"github.com/czcorpus/klogproc/transform"
	"github.com/czcorpus/klogproc/transform/kontext"
)

// kontextLineParser wraps kontext-specific parser into a general form as required
// by core of the klogproc
type kontextLineParser struct {
	lp *kontext.LineParser
}

// ParseLine parses a passed line of a respective log
func (parser *kontextLineParser) ParseLine(s string, lineNum int, localTimezone string) (transform.InputRecord, error) {
	return parser.lp.ParseLine(s, lineNum, localTimezone)
}

func newLineParser(appType string) (LineParser, error) {
	switch appType {
	case "kontext":
		return &kontextLineParser{lp: &kontext.LineParser{}}, nil
	default:
		return nil, fmt.Errorf("Parser not found for application type %s", appType)
	}
}
