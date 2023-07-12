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

package wsserver

import (
	"net"
	"time"

	"klogproc/conversion"
)

// InputRecord represents a raw-parsed version of Word-Sim-Server's access log
type InputRecord struct {
	Action        string  `json:"action"`
	Datetime      string  `json:"datetime"`
	IPAddress     string  `json:"ipAddress"`
	HTTPUserAgent string  `json:"httpUserAgent"`
	Model         string  `json:"model"`
	Corpus        string  `json:"corpus"`
	ProcTime      float64 `json:"procTime"`
	isProcessable bool
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.isProcessable {
		return conversion.ConvertDatetimeStringNoTZ(r.Datetime)
	}
	return time.Time{}
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.IPAddress)
}

func (r *InputRecord) ClusteringClientID() string {
	return conversion.GenerateRandomClusteringID()
}

func (r *InputRecord) ClusterSize() int {
	return 0
}

func (r *InputRecord) SetCluster(size int) {
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (r *InputRecord) GetUserAgent() string {
	return r.HTTPUserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return r.isProcessable
}
