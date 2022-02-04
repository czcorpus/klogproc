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
	"klogproc/conversion"
	"math"
	"time"
)

const (
	minSuspicionLogItems      = 5
	suspiciousStdevRatio      = 0.3
	logRecordsMaxDistanceSecs = 60
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

func (ips *IPProcData) IsSuspicious() bool {
	return ips.Stdev()/ips.mean <= suspiciousStdevRatio && ips.count >= minSuspicionLogItems
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

type Watchdog struct {
	statistics map[string]*IPProcData
	suspicions map[string][]IPProcData
}

func (wd *Watchdog) Register(rec conversion.InputRecord) {
	srec, ok := wd.statistics[rec.GetClientIP().String()]
	if !ok {
		srec = &IPProcData{}
		wd.statistics[rec.GetClientIP().String()] = srec
	}
	// here we use Welford algorithm (https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm)
	if srec.lastAccess.IsZero() {
		srec.firstAccess = rec.GetTime()

	} else {
		if rec.GetTime().Sub(srec.lastAccess) <= logRecordsMaxDistanceSecs*time.Second {
			srec.count++
			timeDist := float64(rec.GetTime().Sub(srec.lastAccess).Milliseconds()) / 1000
			delta := timeDist - srec.mean
			srec.mean += delta / float64(srec.count)
			delta2 := timeDist - srec.mean
			srec.m2 += delta * delta2
		}
		if srec.IsSuspicious() {
			_, ok := wd.suspicions[rec.GetClientIP().String()]
			if !ok {
				wd.suspicions[rec.GetClientIP().String()] = make([]IPProcData, 0, 5)
			}
			wd.suspicions[rec.GetClientIP().String()] = append(wd.suspicions[rec.GetClientIP().String()], *srec)
		}
		if srec.IsSuspicious() || rec.GetTime().Sub(srec.lastAccess) > logRecordsMaxDistanceSecs*time.Second {
			wd.statistics[rec.GetClientIP().String()] = &IPProcData{
				firstAccess: rec.GetTime(),
			}
		}
	}
	srec.lastAccess = rec.GetTime()
}

func (wd *Watchdog) GetSuspiciousRecords() []IPStats {
	ans := make([]IPStats, 0, len(wd.suspicions))
	for ip, recs := range wd.suspicions {
		for _, rec := range recs {
			ans = append(ans, rec.ToIPStats(ip))
		}
	}
	return ans
}

func NewWatchdog() *Watchdog {
	return &Watchdog{
		statistics: make(map[string]*IPProcData),
		suspicions: make(map[string][]IPProcData),
	}
}