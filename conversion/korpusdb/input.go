// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package korpusdb

import (
	"net"
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

type QueryFeat struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
	CI    bool        `json:"ci"`
}

type Query struct {
	Type  string      `json:"type"`
	Feats []QueryFeat `json:"feats"`
}

type Pagination struct {
	From int `json:"from"`
	Size int `json:"size"`
}

type Request struct {
	Feats []string                 `json:"feats"`
	Query Query                    `json:"query"`
	Page  Pagination               `json:"page"`
	Sort  []map[string]interface{} `json:"sort"`
}

// InputRecord is a KorpusDB parsed log record
type InputRecord struct {
	TS      string  `json:"ts"`
	Path    string  `json:"path"`
	Method  string  `json:"method"`
	UserID  string  `json:"userid"`
	IP      string  `json:"ip"`
	Request Request `json:"request"`
}

func (rec *InputRecord) GetTime() time.Time {
	return conversion.ConvertDatetimeStringWithMillisNoTZ(rec.TS)
}

func (rec *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(rec.IP)
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (rec *InputRecord) GetUserAgent() string {
	return ""
}

// IsProcessable returns true if there was no error in reading the record
func (rec *InputRecord) IsProcessable() bool {
	return true
}
