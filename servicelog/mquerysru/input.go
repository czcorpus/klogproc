// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package mquerysru

import (
	"klogproc/servicelog"
	"net"
	"time"
)

// InputRecord represents a raw-parsed version of masm query log
type InputRecord struct {
	// gokit logging middleware log events
	Level        string  `json:"level"`
	Time         string  `json:"time"`
	Latency      float64 `json:"latency"`
	ClientIP     string  `json:"clientIP"`
	Method       string  `json:"method"`
	Status       int     `json:"status"`
	ErrorMessage string  `json:"errorMessage"`
	BodySize     int     `json:"bodySize"`
	Path         string  `json:"path"`

	Version           string `json:"version"`
	Operation         string `json:"operation"`
	RecordXMLEscaping string `json:"recordXMLEscaping"`
	RecordPacking     string `json:"recordPacking"`

	Args map[string]interface{} `json:"args"`
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.Time[len(r.Time)-1] == 'Z' {
		return servicelog.ConvertDatetimeString(r.Time[:len(r.Time)-1] + "+00:00")
	}
	return servicelog.ConvertDatetimeString(r.Time)
}

func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.ClientIP)
}

func (rec *InputRecord) ClusteringClientID() string {
	return servicelog.GenerateRandomClusteringID()
}

func (rec *InputRecord) ClusterSize() int {
	return 0
}

func (rec *InputRecord) SetCluster(size int) {
}

func (r *InputRecord) GetUserAgent() string {
	return ""
}

func (r *InputRecord) IsProcessable() bool {
	// process only http requests
	return len(r.Method) > 0
}

func (rec *InputRecord) IsSuspicious() bool {
	return false
}

func (rec *InputRecord) IsQuery() bool {
	return rec.Operation == "searchRetrieve"
}
