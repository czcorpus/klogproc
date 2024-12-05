// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package treqapi

import (
	"klogproc/servicelog"
	"net"
	"time"
)

/*
{"time":"2024-12-05T10:19:36+01:00","ip":"31.0.2.157","user_id":"1531","query":"zvládnout","from":"cs",
"to":"pl","asc":false,"regex":false,"lemma":true,"multiword":false,"ci":true,"order":"perc",
"pkgs":["CORE","EUROPARL","PRESSEUROP","SUBTITLES"]}
*/

// InputRecord is a Treq parsed log record
type InputRecord struct {
	Time      string   `json:"time"`
	IP        string   `json:"ip"`
	UserID    string   `json:"user_id"`
	Query     string   `json:"query"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	Asc       bool     `json:"asc"`
	Regex     bool     `json:"regex"`
	Lemma     bool     `json:"lemma"`
	Multiword bool     `json:"multiword"`
	CI        bool     `json:"ci"`
	Order     string   `json:"order"`
	Pkgs      []string `json:"pkgs"`
}

func (rec *InputRecord) GetTime() time.Time {
	return servicelog.ConvertDatetimeString(rec.Time)
}

func (rec *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(rec.IP)
}

func (rec *InputRecord) ClusteringClientID() string {
	return servicelog.GenerateRandomClusteringID()
}

func (rec *InputRecord) ClusterSize() int {
	return 0
}

func (rec *InputRecord) SetCluster(size int) {
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (rec *InputRecord) GetUserAgent() string {
	return ""
}

// IsProcessable returns true if there was no error in reading the record
func (rec *InputRecord) IsProcessable() bool {
	return true
}

func (rec *InputRecord) IsSuspicious() bool {
	return false
}
