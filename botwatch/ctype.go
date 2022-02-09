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

package botwatch

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"klogproc/botwatch/redisdb"
	"klogproc/common"
	"klogproc/conversion"
)

type BotInfo struct {
	Title   string   `json:"title"`
	Match   []string `json:"match"`
	Example string   `json:"example"`
}

type BotsAndMonitors struct {
	Bots        []BotInfo `json:"bots"`
	Monitors    []BotInfo `json:"monitors"`
	IPBlacklist []string  `json:"ipBlacklist"`
}

func searchMatchingDef(rec conversion.InputRecord, defs []BotInfo) bool {
	for _, item := range defs {
		match := true
		for _, m := range item.Match {
			match = match && strings.Contains(strings.ToLower(rec.GetUserAgent()), m)
			if !match {
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func importIPList(data []string) ([]net.IP, error) {
	ans := make([]net.IP, len(data))
	for i, ips := range data {
		ip := net.ParseIP(ips)
		if ip == nil {
			return ans, fmt.Errorf("cannot parse configured IP %s", ips)
		}
		ans[i] = ip
	}
	return ans, nil
}

type ClientTypeAnalyzer struct {
	bots        []BotInfo
	monitors    []BotInfo
	iPBlacklist []net.IP
	botWatch    *Watchdog
	outDB       *redisdb.RedisWriter
}

func (cta *ClientTypeAnalyzer) AgentIsMonitor(rec conversion.InputRecord) bool {
	return searchMatchingDef(rec, cta.monitors)
}

func (cta *ClientTypeAnalyzer) AgentIsBot(rec conversion.InputRecord) bool {
	return searchMatchingDef(rec, cta.bots)
}

func (cta *ClientTypeAnalyzer) HasBlacklistedIP(rec conversion.InputRecord) bool {
	for _, ip := range cta.iPBlacklist {
		if rec.GetClientIP().Equal(ip) {
			return true
		}
	}
	return false
}

func (cta *ClientTypeAnalyzer) Add(rec conversion.InputRecord) {
	cta.botWatch.Add(rec)
}

func (cta *ClientTypeAnalyzer) ResetBotCandidates() {
	cta.botWatch.ResetBotCandidates()
}

func (cta *ClientTypeAnalyzer) GetBotCandidates() []IPStats {
	return cta.botWatch.GetSuspiciousRecords()
}

func (cta *ClientTypeAnalyzer) StoreBotCandidates() {
	for _, rec := range cta.GetBotCandidates() {
		err := cta.outDB.WriteReport(rec.IP, &redisdb.ActivityReport{
			NumRequests:    rec.Count,
			TimeWindowSecs: cta.botWatch.conf.WatchedTimeWindowSecs,
			Mean:           rec.Mean,
			Stdev:          rec.Stdev,
			Created:        rec.LastRequest,
		})
		if err != nil {
			fmt.Print("ERROR: error writing botwatch report to Redis: ", err)
			// TODO trigger e-mail alarm
		}
	}
}

func (cta *ClientTypeAnalyzer) Close() {
	cta.botWatch.Close()
}

func NewClientTypeAnalyzer(cfg BotDetectionConf) (*ClientTypeAnalyzer, error) {
	conf := new(BotsAndMonitors)
	var listIP []net.IP

	if cfg.BotDefsPath != "" {
		log.Printf("INFO: using bot definitions from resource %s", cfg.BotDefsPath)
		rawData, err := common.LoadSupportedResource(cfg.BotDefsPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(rawData, conf)
		if err != nil {
			return nil, err
		}
		for i, mList := range conf.Bots {
			for j, m := range mList.Match {
				conf.Bots[i].Match[j] = strings.ToLower(m)
			}
		}
		for i, mList := range conf.Monitors {
			for j, m := range mList.Match {
				conf.Monitors[i].Match[j] = strings.ToLower(m)
			}
		}
		listIP, err = importIPList(conf.IPBlacklist)
		if err != nil {
			return nil, err
		}

	} else {
		log.Print("WARNING: no bots configuration provided (botDefsPath)")
	}
	log.Printf("INFO: bot defs: %d, monitors defs: %d, blacklisted IPs: %d", len(conf.Bots), len(conf.Monitors), len(listIP))

	outDB := redisdb.NewRedisWriter(cfg.Redis)

	return &ClientTypeAnalyzer{
			bots:        conf.Bots,
			monitors:    conf.Monitors,
			iPBlacklist: listIP,
			botWatch:    NewWatchdog(cfg),
			outDB:       outDB,
		},
		nil
}
