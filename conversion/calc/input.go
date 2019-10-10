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

package calc

import (
	"net"
	"time"

	"github.com/czcorpus/klogproc/conversion"
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
	return conversion.ConvertDatetimeString(r.TS)
}

// GetClientIP returns a normalized IP address info
func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.ClientIP)
}

// AgentIsLoggable returns true if the record should be stored.
// Otherwise (bots, static files access, some operations) it
// returns false and klogproc ignores such record.
func (r *InputRecord) AgentIsLoggable() bool {
	return true // TODO
}