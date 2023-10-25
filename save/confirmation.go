// Copyright 2021 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 202 Institute of the Czech National Corpus,
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

package save

import (
	"fmt"
	"klogproc/servicelog"
)

type ConfirmMsg struct {
	FilePath string
	Position servicelog.LogRange
	Error    error
}

func (cm ConfirmMsg) String() string {
	return fmt.Sprintf("ConfirmMsg{FilePath: %v, Position: %v, Error: %v}", cm.FilePath, cm.Position, cm.Error)
}

// --------------------

type IgnoredItemMsg struct {
	FilePath string
	Position servicelog.LogRange
}

func (iim IgnoredItemMsg) String() string {
	return fmt.Sprintf("IgnoredItemMsg{FilePath: %v, Position: %v}", iim.FilePath, iim.Position)
}

func NewIgnoredItemMsg(filePath string, position servicelog.LogRange) IgnoredItemMsg {
	newPos := position
	newPos.Written = true
	return IgnoredItemMsg{FilePath: filePath, Position: newPos}
}
