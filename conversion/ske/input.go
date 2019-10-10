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

package ske

import (
	"net"
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

// Request is a simple representation of
// HTTP request metadata used in KonText logging
type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
	HTTPRemoteAddr   string `json:"HTTP_REMOTE_ADDR"`
	RemoteAddr       string `json:"REMOTE_ADDR"`
}

// InputRecord represents a raw-parsed version of SkE's access log
type InputRecord struct {
	Action     string
	Corpus     string
	Subcorpus  string
	Datetime   string
	User       string
	Request    Request
	isLoggable bool
	ProcTime   float32
	// TODO
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	return conversion.ConvertAccessLogDatetimeString(r.Datetime)
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.Request.RemoteAddr)
}

// AgentIsLoggable returns true if the record should be stored.
// Otherwise (bots, static files access, some operations) it
// returns false and klogproc ignores such record.
func (r *InputRecord) AgentIsLoggable() bool {
	return r.isLoggable
}