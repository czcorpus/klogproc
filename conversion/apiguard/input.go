// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package apiguard

import (
	"net"
	"time"

	"klogproc/conversion"
)

// InputRecord represents a parsed KonText record
type InputRecord struct {
	Type       string  `json:"type"`
	Level      string  `json:"level"`
	AccessLog  bool    `json:"accessLog"`
	Service    string  `json:"service"`
	ProcTime   float64 `json:"procTime"`
	IsCached   bool    `json:"isCached"`
	IsIndirect bool    `json:"isIndirect"`
	UserID     *int    `json:"userId"`
	Time       string  `json:"time"`
	IPAddress  string  `json:"ipAddress"`
	UserAgent  string  `json:"userAgent"`
}

// GetTime returns record's time as a Golang's Time
// instance. Please note that the value is truncated
// to seconds.
func (rec *InputRecord) GetTime() time.Time {
	return conversion.ConvertDatetimeString(rec.Time)
}

// GetClientIP returns a client IP no matter in which
// part of the record it was found
func (rec *InputRecord) GetClientIP() net.IP {
	if rec.IPAddress != "" {
		return net.ParseIP(rec.IPAddress)
	}
	return make([]byte, 0)
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (rec *InputRecord) GetUserAgent() string {
	return rec.UserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (rec *InputRecord) IsProcessable() bool {
	return rec.AccessLog
}