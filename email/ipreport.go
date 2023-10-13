// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

// fileselect functions are used to find proper KonText application log files
// based on logs processed so far. Please note that in recent KonText and
// Klogproc versions this is rather a fallback/offline functionality.

package email

import (
	"fmt"
	"strings"
)

type TR struct {
	tbody *TBody
}

func (tr *TR) AddTH(text string, style string) *TR {
	if style != "" {
		tr.tbody.table.writer.WriteString(fmt.Sprintf("<th style=\"%s\">%s</th>", style, text))

	} else {
		tr.tbody.table.writer.WriteString(fmt.Sprintf("<th>%s</th>", text))
	}
	return tr
}

func (tr *TR) AddTD(text string, style string) *TR {
	if style != "" {
		tr.tbody.table.writer.WriteString(fmt.Sprintf("<td style=\"%s\">%s</td>", style, text))

	} else {
		tr.tbody.table.writer.WriteString(fmt.Sprintf("<td>%s</td>", text))
	}
	return tr
}

func (tr *TR) Close() *TBody {
	tr.tbody.table.writer.WriteString("</tr>")
	return tr.tbody
}

type TBody struct {
	table *Table
}

func (t *TBody) AddTR() *TR {
	t.table.writer.WriteString("<tr>")
	return &TR{tbody: t}
}

func (t *TBody) Close() *Table {
	t.table.writer.WriteString("</tbody>")
	return t.table
}

type Table struct {
	writer *strings.Builder
}

func (t *Table) Init(style string) *Table {
	t.writer = &strings.Builder{}
	t.writer.WriteString(fmt.Sprintf("<table style=\"%s\">", style))
	return t
}

func (t *Table) AddBody() *TBody {
	t.writer.WriteString("<tbody>")
	return &TBody{table: t}
}

func (t *Table) Close() {
	t.writer.WriteString("</table>")
}

func (t *Table) String() string {
	return t.writer.String()
}
