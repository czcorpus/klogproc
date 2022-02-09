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
	"klogproc/conversion"
	"sync"
	"time"
)

type Watchdog struct {
	statistics     map[string]*IPProcData
	suspicions     map[string]IPProcData
	conf           BotDetectionConf
	onlineAnalysis chan conversion.InputRecord
	mutex          sync.Mutex
}

func (wd *Watchdog) maxLogRecordsDistance() time.Duration {
	return time.Duration(wd.conf.WatchedTimeWindowSecs/wd.conf.NumRequestsThreshold) * time.Second
}

func (wd *Watchdog) Add(rec conversion.InputRecord) {
	wd.onlineAnalysis <- rec
}

func (wd *Watchdog) Close() {
	close(wd.onlineAnalysis)
}

func (wd *Watchdog) ResetAll() {
	wd.mutex.Lock()
	wd.statistics = make(map[string]*IPProcData)
	wd.suspicions = make(map[string]IPProcData)
	wd.mutex.Unlock()
}

func (wd *Watchdog) ResetBotCandidates() {
	wd.mutex.Lock()
	wd.suspicions = make(map[string]IPProcData)
	wd.mutex.Unlock()
}

func (wd *Watchdog) Conf() BotDetectionConf {
	return wd.conf
}

func (wd *Watchdog) analyze(rec conversion.InputRecord) {
	srec, ok := wd.statistics[rec.GetClientIP().String()]
	if !ok {
		srec = &IPProcData{}
		wd.statistics[rec.GetClientIP().String()] = srec
	}
	// here we use Welford algorithm for online variance calculation
	// more info: (https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm)
	if srec.lastAccess.IsZero() {
		srec.firstAccess = rec.GetTime()

	} else {
		if rec.GetTime().Sub(srec.lastAccess) <= wd.maxLogRecordsDistance() {
			srec.count++
			timeDist := float64(rec.GetTime().Sub(srec.lastAccess).Milliseconds()) / 1000
			delta := timeDist - srec.mean
			srec.mean += delta / float64(srec.count)
			delta2 := timeDist - srec.mean
			srec.m2 += delta * delta2
		}
		if srec.IsSuspicious(wd.conf) {
			prev, ok := wd.suspicions[rec.GetClientIP().String()]
			if !ok || srec.ReqPerSecod() > prev.ReqPerSecod() {
				wd.suspicions[rec.GetClientIP().String()] = *srec
			}
		}
		if srec.IsSuspicious(wd.conf) || rec.GetTime().Sub(srec.firstAccess) > time.Duration(wd.conf.WatchedTimeWindowSecs)*time.Second {
			wd.statistics[rec.GetClientIP().String()] = &IPProcData{
				firstAccess: rec.GetTime(),
			}
		}
	}
	srec.lastAccess = rec.GetTime()
}

func (wd *Watchdog) GetSuspiciousRecords() []IPStats {
	wd.mutex.Lock()
	defer wd.mutex.Unlock()
	ans := make([]IPStats, 0, len(wd.suspicions))
	for ip, rec := range wd.suspicions {
		ans = append(ans, rec.ToIPStats(ip))
	}
	return ans
}

func NewWatchdog(conf BotDetectionConf) *Watchdog {
	analysis := make(chan conversion.InputRecord)
	wd := &Watchdog{
		statistics:     make(map[string]*IPProcData),
		suspicions:     make(map[string]IPProcData),
		conf:           conf,
		onlineAnalysis: analysis,
	}
	go func() {
		for item := range analysis {
			wd.analyze(item)
		}
	}()
	return wd
}
