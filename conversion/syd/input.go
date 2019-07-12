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

package syd

import (
	"net"
	"time"
)

type InputRecord struct {
	Datetime  string
	IPAddress string
	UserID    int
	KeyReq    string
	KeyUsed   string
	Key       string
	Ltool     string
	RunScript string
}

func (r *InputRecord) GetTime() time.Time {
	if t, err := time.Parse("2006-01-02 15:04:05-07:00", r.Datetime); err == nil {
		return t
	}
	return time.Time{}
}

func (r *InputRecord) GetClientIP() net.IP {
	return net.ParseIP(r.IPAddress)
}

func (r *InputRecord) AgentIsLoggable() bool {
	return true
}
