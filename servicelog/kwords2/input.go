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

package kwords2

import (
	"klogproc/servicelog"
	"net"
	"time"
)

/*
{
	"time": "2023-10-30T14:53:57.247221",
	"client": ["195.113.53.123", 0],
	"headers": {
		"host": "kwords2.korpus.cz",
		"x-forwarded-for": "195.113.53.123",
		"connection": "close",
		"content-length": "250",
		"user-agent": "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/118.0",
		"accept": "application/json, text/plain, **",
		 "accept-language": "cs,sk;q=0.8,en-US;q=0.5,en;q=0.3",
		 "accept-encoding": "gzip, deflate, br",
		 "content-type": "application/json",
		 "origin": "https://kwords2.korpus.cz",
		 "referer": "https://kwords2.korpus.cz/",
		 "sec-fetch-dest": "empty",
		 "sec-fetch-mode": "cors",
		 "sec-fetch-site": "same-origin",
		 "cookie": "_ga=GA1.2.1828084022.1695714932; _ga_6KFK7LV8PZ},...",
		"method": "POST",
		"path": "/api/keywords",
		"path_params": {},
		"query_params": {},
	"body": {
		"textId": "3b6d97d8c2f3143462cea7c6d01e97cf",
		"refCorpus": "syn2020ud",
		"attrs": ["word"],
		"stopList": [{"attr": "upos", "values": ["ADP", "CCONJ", "SCONJ", "NUM", "PUNCT"]}],
		"statTest": "log-likelihood",
		"level": 0.05,
		"effectMetric": "din",
		"minFreq": 3,
		"percent": 100},
	"exception": null,
	"response": {"taskId": "2ddaf235c0aa80203441e5c068266f6f"}}
*/

type Body struct {
	TextID       string   `json:"textId"`
	RefCorpus    string   `json:"refCorpus"`
	Attrs        []string `json:"attrs"`
	Level        float64  `json:"level"`
	EffectMetric string   `json:"effectMetric"`

	// MinFreq
	// note to any type: there are either strings or ints in logs
	MinFreq any `json:"minFreq"`

	// Percent
	// note to any type: there are either strings or ints in logs
	Percent any `json:"percent"`

	// the following two items are used in case of text import

	Text string `json:"text"`
	Lang string `json:"lang"`
}

type Headers struct {
	UserAgent string `json:"user-agent"`
}

// InputRecord is a Kwords parsed log record
type InputRecord struct {
	Time      string  `json:"time"`
	Client    [2]any  `json:"client"`
	Headers   Headers `json:"headers"`
	Path      string  `json:"path"`
	Body      Body    `json:"body"`
	UserID    int     `json:"userId"`
	Exception string  `json:"exception,omitempty"`
}

func (rec *InputRecord) GetTime() time.Time {
	return servicelog.ConvertDatetimeStringWithMillisNoTZ(rec.Time)
}

func (rec *InputRecord) GetClientIP() net.IP {
	v, ok := rec.Client[0].(string)
	if !ok {
		return net.IP{}
	}
	return net.ParseIP(v)
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
