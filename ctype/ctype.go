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

package ctype

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/czcorpus/klogproc/config"
	"github.com/czcorpus/klogproc/conversion"
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
			match = match && strings.Index(strings.ToLower(rec.GetUserAgent()), m) > -1
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
			return ans, fmt.Errorf("Cannot parse configured IP %s", ips)
		}
		ans[i] = ip
	}
	return ans, nil
}

type ClientTypeAnalyzer struct {
	bots        []BotInfo
	monitors    []BotInfo
	iPBlacklist []net.IP
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

func LoadFromResource(path string) (*ClientTypeAnalyzer, error) {
	rawData, err := config.LoadSupportedResource(path)
	conf := new(BotsAndMonitors)
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
	listIP, err := importIPList(conf.IPBlacklist)
	if err != nil {
		return nil, err
	}
	log.Printf("INFO: bot defs: %d, monitors defs: %d, blacklisted IPs: %d", len(conf.Bots), len(conf.Monitors), len(listIP))
	return &ClientTypeAnalyzer{bots: conf.Bots, monitors: conf.Monitors, iPBlacklist: listIP}, nil
}
