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

package mapka3

import (
	"net"
	"time"

	"klogproc/conversion"
)

// InputRecord represents a raw-parsed version of MAPKA's access log
// the source looks like this:
// {"message":"Marker","context":{},"level":200,"level_name":"INFO",
// "channel":"access","datetime":"2023-07-04T17:18:22.294828+02:00",
// "extra":{"session_selector":"aa7e3e2a322a","user_id":"4321","url":"/markers",
// "ip":"::1","http_method":"POST","server":"localhost","referrer":"http://localhost:8083/"}}
type InputRecord struct {
	Message       string         `json:"message"`
	Context       map[string]any `json:"context"`
	Level         int            `json:"level"`
	LevelName     string         `json:"level_name"`
	Channel       string         `json:"access"`
	Datetime      string         `json:"datetime"`
	Extra         Extra          `json:"extra"`
	isProcessable bool
}

type Extra struct {
	SessionSelector string `json:"session_selector"`
	UserID          string `json:"user_id"`
	URL             string `json:"url"`
	IP              string `json:"ip"`
	HTTPMethod      string `json:"http_method"`
	Server          string `json:"server"`
	Referrer        string `json:"referrer"`
}

// GetTime returns a normalized log date and time information
func (r *InputRecord) GetTime() time.Time {
	return conversion.ConvertDatetimeStringWithMillis(r.Datetime)
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	v := net.ParseIP(r.Extra.IP)
	if v != nil {
		return v
	}
	return net.IPv4zero
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (r *InputRecord) GetUserAgent() string {
	return ""
}

// IsProcessable returns true if there was no error in reading the record
func (r *InputRecord) IsProcessable() bool {
	return true
}
