// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package wag06

import (
	"net"
	"time"

	"klogproc/conversion"
)

// Request is a simple representation of
// HTTP request metadata used in WaG logging
type Request struct {
	HTTPForwardedFor string
	HTTPUserAgent    string
	HTTPRemoteAddr   string
	RemoteAddr       string
	Referer          string // the misspelling is intentional
}

// InputRecord represents a raw-parsed version of WaG access log
type InputRecord struct {
	Action              string
	QueryType           string
	Lang1               string
	Lang2               string
	Queries             []string
	Datetime            string
	Request             Request
	ProcTime            float32
	isProcessable       bool
	IsMobileClient      bool
	HasPosSpecification bool
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.isProcessable {
		return conversion.ConvertAccessLogDatetimeString(r.Datetime)
	}
	return time.Time{}
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.Request.RemoteAddr)
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (r *InputRecord) GetUserAgent() string {
	return r.Request.HTTPUserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return r.isProcessable
}
