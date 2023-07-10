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

package shiny

import (
	"net"
	"time"

	"klogproc/conversion"
)

type User struct {
	ID        int    `json:"id"`
	User      string `json:"user"`
	FirstName string `json:"firstName"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
}

type InputRecord struct {
	TS        string `json:"ts"`
	ClientIP  string `json:"clientIP"`
	User      User   `json:"user"`
	Lang      string `json:"lang"`
	UserAgent string `json:"userAgent"`
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	if r.TS[len(r.TS)-1] == 'Z' { // UTC time
		return conversion.ConvertDatetimeStringNoTZ(r.TS[:len(r.TS)-1])
	}
	return conversion.ConvertDatetimeString(r.TS)
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.ClientIP)
}

func (rec *InputRecord) ClusteringClientID() string {
	return conversion.GenerateRandomClusteringID()
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (rec *InputRecord) GetUserAgent() string {
	return ""
}

// IsProcessable returns true if there was no error in reading the record
func (rec *InputRecord) IsProcessable() bool {
	return true
}
