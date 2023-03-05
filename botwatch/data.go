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

package botwatch

import (
	"encoding/json"
	"math"
	"time"
)

type IPStats struct {
	IP           string  `json:"ip"`
	Mean         float64 `json:"mean"`
	Stdev        float64 `json:"stdev"`
	Count        int     `json:"count"`
	FirstRequest string  `json:"firstRequest"`
	LastRequest  string  `json:"lastRequest"`
}

func (r *IPStats) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// --------------

type IPProcData struct {
	count       int
	mean        float64
	m2          float64
	firstAccess time.Time
	lastAccess  time.Time
}

func (ips *IPProcData) Variance() float64 {
	if ips.count == 0 {
		return 0
	}
	return ips.m2 / float64(ips.count)
}

func (ips *IPProcData) Stdev() float64 {
	return math.Sqrt(ips.Variance())
}

func (ips *IPProcData) ReqPerSecod() float64 {
	return float64(ips.count) / ips.lastAccess.Sub(ips.lastAccess).Seconds()
}

func (ips *IPProcData) IsSuspicious(conf BotDetectionConf) bool {
	return ips.Stdev()/ips.mean <= conf.RSDThreshold && ips.count >= conf.NumRequestsThreshold
}

func (ips *IPProcData) ToIPStats(ip string) IPStats {
	return IPStats{
		IP:           ip,
		Mean:         ips.mean,
		Stdev:        ips.Stdev(),
		Count:        ips.count,
		FirstRequest: ips.firstAccess.Format(time.RFC3339),
		LastRequest:  ips.lastAccess.Format(time.RFC3339),
	}
}
