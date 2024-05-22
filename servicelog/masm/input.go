// Copyright 2022 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2022 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2022 Institute of the Czech National Corpus,
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

package masm

import (
	"klogproc/servicelog"
	"net"
	"time"
)

// InputRecord represents a raw-parsed version of masm query log
type InputRecord struct {
	Level          string   `json:"level"`
	Time           string   `json:"time"`
	Message        string   `json:"message"`
	IsQuery        bool     `json:"isQuery,omitempty"`
	Corpus         string   `json:"corpus,omitempty"`
	AlignedCorpora []string `json:"alignedCorpora,omitempty"`
	IsAutocomplete bool     `json:"isAutocomplete,omitempty"`
	IsCached       bool     `json:"isCached,omitempty"`
	ProcTimeSecs   float64  `json:"procTimeSecs,omitempty"`
	Error          string   `json:"error,omitempty"`
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.Time[len(r.Time)-1] == 'Z' {
		return servicelog.ConvertDatetimeString(r.Time[:len(r.Time)-1] + "+00:00")
	}
	return servicelog.ConvertDatetimeString(r.Time)
}

func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP("0.0.0.0")
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
	return true
}

func (rec *InputRecord) IsSuspicious() bool {
	return false
}

func (r *InputRecord) ExportError() *servicelog.ErrorRecord {
	if r.Error != "" {
		return &servicelog.ErrorRecord{
			Name: r.Error,
		}
	}
	return nil
}
