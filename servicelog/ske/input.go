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
	"klogproc/servicelog"
	"net"
	"time"
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
	Action        string
	Corpus        string
	Subcorpus     string
	Datetime      string
	User          string
	Request       Request
	ProcTime      float64
	isProcessable bool
	// TODO
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.isProcessable {
		return servicelog.ConvertAccessLogDatetimeString(r.Datetime)
	}
	return time.Time{}
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.Request.RemoteAddr)
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
func (r *InputRecord) GetUserAgent() string {
	return r.Request.HTTPUserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return r.isProcessable
}

func (rec *InputRecord) IsSuspicious() bool {
	return false
}
