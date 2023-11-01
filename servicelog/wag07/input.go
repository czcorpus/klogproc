// Copyright 2021 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2021 Institute of the Czech National Corpus,
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

package wag07

import (
	"klogproc/servicelog"
	"net"
	"time"
)

// Request is a simple representation of
// HTTP request metadata used in WaG logging
type Request struct {
	HTTPForwardedFor string `json:"httpForwardedFor"`
	HTTPUserAgent    string `json:"userAgent"`
	Origin           string `json:"origin"`
	Referer          string `json:"referer"` // the misspelling is intentional
}

// InputRecord represents a raw-parsed version of WaG query log
type InputRecord struct {
	Level               string  `json:"level"`
	Message             string  `json:"message"`
	UserID              int     `json:"userId"`
	Action              string  `json:"action"`
	QueryType           string  `json:"queryType"`
	Request             Request `json:"request"`
	Lang1               string  `json:"lang1"`
	Lang2               string  `json:"lang2"`
	Timestamp           string  `json:"timestamp"`
	isProcessable       bool
	IsMobileClient      bool `json:"isMobileClient"`
	HasMatch            bool `json:"hasMatch"`
	IsQuery             bool `json:"isQuery"`
	HasPosSpecification bool `json:"hasPosSpecification"`
}

func (r *InputRecord) ShouldBeAnalyzed() bool {
	return r.Action == "search" || r.Action == "compare" || r.Action == "translate"
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.isProcessable {
		if r.Timestamp[len(r.Timestamp)-1] == 'Z' {
			return servicelog.ConvertDatetimeStringWithMillisNoTZ(r.Timestamp[:len(r.Timestamp)-1] + "000")
		}
		return servicelog.ConvertDatetimeStringWithMillisNoTZ(r.Timestamp + "000")
	}
	return time.Time{}
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.Request.Origin)
}

func (r *InputRecord) ClusteringClientID() string {
	return r.GetClientIP().String()
}

func (r *InputRecord) ClusterSize() int {
	return 0
}

func (r *InputRecord) SetCluster(size int) {
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (r *InputRecord) GetUserAgent() string {
	return r.Request.HTTPUserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return r.isProcessable
}

func (rec *InputRecord) IsSuspicious() bool {
	return rec.IsQuery && !rec.HasMatch
}
