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

package mapka

import (
	"net"
	"time"

	"klogproc/conversion"
)

// Request is a simple representation of
// HTTP request metadata used in KonText logging
type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
	HTTPRemoteAddr   string `json:"HTTP_REMOTE_ADDR"`
	RemoteAddr       string `json:"REMOTE_ADDR"`
}

// RequestParams is a mix of some significant params of watched requests
type RequestParams struct {
	CardType    *string `json:"cardType"`
	CardFolder  *string `json:"cardFolder"`
	OverlayFile *string `json:"overlayFile"`
}

// InputRecord represents a raw-parsed version of MAPKA's access log
type InputRecord struct {
	Action        string
	Path          string
	Datetime      string
	Request       *Request
	Params        *RequestParams `json:"params"`
	ProcTime      float32
	isProcessable bool
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
	if r.Request != nil {
		return net.ParseIP(r.Request.RemoteAddr)
	}
	return net.IPv4zero
}

func (rec *InputRecord) ClusteringClientID() string {
	return conversion.GenerateRandomClusteringID()
}

func (rec *InputRecord) ClusterSize() int {
	return 0
}

func (rec *InputRecord) SetCluster(size int) {
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (r *InputRecord) GetUserAgent() string {
	if r.Request != nil {
		return r.Request.HTTPUserAgent
	}
	return ""
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return r.isProcessable
}
