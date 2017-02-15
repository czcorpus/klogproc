// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
	HTTPRemoteAddr   string `json:"HTTP_REMOTE_ADDR"`
}

type LogRecord struct {
	UserID   int               `json:"user_id"`
	ProcTime float32           `json:"proc_time"`
	Date     string            `json:"date"`
	Action   string            `json:"action"`
	Request  Request           `json:"request"`
	Params   map[string]string `json:"params"`
}

func (rec *LogRecord) GetTime() time.Time {
	if t, err := time.Parse("2006-01-02 15:04:05", rec.Date); err == nil {
		return t
	}
	return time.Time{}
}

func (rec *LogRecord) GetClientIP() net.IP {
	if rec.Request.HTTPForwardedFor != "" {
		return net.ParseIP(rec.Request.HTTPForwardedFor)

	} else if rec.Request.HTTPUserAgent != "" {
		return net.ParseIP(rec.Request.HTTPRemoteAddr)
	}
	return make([]byte, 0)
}

func (rec *LogRecord) AgentIsBot() bool {
	agentStr := strings.ToLower(rec.Request.HTTPUserAgent)
	return strings.Index(agentStr, "googlebot") > -1 ||
		strings.Index(agentStr, "ahrefsbot") > -1 ||
		strings.Index(agentStr, "yandexbot") > -1 ||
		strings.Index(agentStr, "yahoo") > -1 && strings.Index(agentStr, "slurp") > -1 ||
		strings.Index(agentStr, "baiduspider") > -1 ||
		strings.Index(agentStr, "seznambot") > -1 ||
		strings.Index(agentStr, "bingbot") > -1 ||
		strings.Index(agentStr, "megaindex.ru") > -1
}

func (rec *LogRecord) AgentIsMonitor() bool {
	agentStr := strings.ToLower(rec.Request.HTTPUserAgent)
	return strings.Index(agentStr, "python-urllib/2.7") > -1 ||
		strings.Index(agentStr, "zabbix-test") > -1
}

func (rec *LogRecord) AgentIsLoggable() bool {
	return !rec.AgentIsBot() && !rec.AgentIsMonitor()
}

type LogInterceptor interface {
	ProcItem(appType string, record *LogRecord)
}

func parseRawLine(s string) string {
	reg := regexp.MustCompile("^.+\\sINFO:\\s+(\\{.+)$")
	srch := reg.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func getLineType(s string) string {
	reg := regexp.MustCompile("^.+\\s([A-Z]+):\\s+.+$")
	srch := reg.FindStringSubmatch(s)
	if srch != nil {
		return srch[1]
	}
	return ""
}

func importDatetimeString(dateStr string, localTimezone string) string {
	rg := regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2})\\s([012]\\d:[0-5]\\d:[0-5]\\d\\.\\d+)")
	srch := rg.FindStringSubmatch(dateStr)
	if len(srch) > 0 {
		return fmt.Sprintf("%sT%s%s", srch[1], srch[2], localTimezone)
	}
	return ""
}

func NewParser(path string, geoIPPath string, localTimezone string) *Parser {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(f)
	return &Parser{fr: sc, localTimezone: localTimezone}
}

type Parser struct {
	fr            *bufio.Scanner
	localTimezone string
}

func (p *Parser) parseLine(s string, recType string, proc LogInterceptor) {
	jsonLine := parseRawLine(s)
	//fmt.Println("LINE: ", jsonLine)
	if jsonLine != "" {
		var record LogRecord
		err := json.Unmarshal([]byte(jsonLine), &record)
		if err == nil {
			record.Date = importDatetimeString(record.Date, p.localTimezone)
			proc.ProcItem(recType, &record)
		}

	} else if tp := getLineType(s); tp == "QUERY" {
		log.Printf("Failed to process QUERY entry: %s", s)
	}
}

func (p *Parser) Parse(fromTimestamp int64, recType string, proc LogInterceptor) {
	fmt.Println("parsing from timestamp: ", fromTimestamp)
	for p.fr.Scan() {
		p.parseLine(p.fr.Text(), recType, proc)
	}
}
