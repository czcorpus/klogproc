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

// ------------------------------------------------------------

// Request is a simple representation of
// HTTP request metadata used in KonText logging
type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
	HTTPRemoteAddr   string `json:"HTTP_REMOTE_ADDR"`
}

// ------------------------------------------------------------

// LogRecord represents a parsed KonText record
type LogRecord struct {
	UserID   int                    `json:"user_id"`
	ProcTime float32                `json:"proc_time"`
	Date     string                 `json:"date"`
	Action   string                 `json:"action"`
	Request  Request                `json:"request"`
	Params   map[string]interface{} `json:"params"`
	PID      int                    `json:"pid"`
	Settings map[string]interface{} `json:"settings"`
}

// GetTime returns record's time as a Golang's Time
// instance. Please note that the value is truncated
// to seconds.
func (rec *LogRecord) GetTime() time.Time {
	p := regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2}T[012]\\d:[0-5]\\d:[0-5]\\d)\\.\\d+")
	srch := p.FindStringSubmatch(rec.Date)
	if srch != nil {
		if t, err := time.Parse("2006-01-02T15:04:05", srch[1]); err == nil {
			return t
		}
	}
	return time.Time{}
}

// GetClientIP returns a client IP no matter in which
// part of the record it was found
// (e.g. HTTP_USER_AGENT vs. HTTP_FORWARDED_FOR)
func (rec *LogRecord) GetClientIP() net.IP {
	if rec.Request.HTTPForwardedFor != "" {
		return net.ParseIP(rec.Request.HTTPForwardedFor)

	} else if rec.Request.HTTPUserAgent != "" {
		return net.ParseIP(rec.Request.HTTPRemoteAddr)
	}
	return make([]byte, 0)
}

// AgentIsBot returns true if user agent information suggests
// that the client is not human. The rules are currently
// hardcoded and quite simple.
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

// AgentIsMonitor returns true if user agent information
// matches one of "bots" used by the Instatute Czech National Corpus
// to monitor service availability. The rules are currently
// hardcoded.
func (rec *LogRecord) AgentIsMonitor() bool {
	agentStr := strings.ToLower(rec.Request.HTTPUserAgent)
	return strings.Index(agentStr, "python-urllib/2.7") > -1 ||
		strings.Index(agentStr, "zabbix-test") > -1
}

// AgentIsLoggable returns true if the current record
// is determined to be saved (we ignore bots, monitors etc.).
func (rec *LogRecord) AgentIsLoggable() bool {
	return !rec.AgentIsBot() && !rec.AgentIsMonitor()
}

func (rec *LogRecord) GetStringParam(name string) string {
	switch v := rec.Params[name].(type) {
	case string:
		return v
	}
	return ""
}

func (rec *LogRecord) GetIntParam(name string) int {
	switch v := rec.Params[name].(type) {
	case int:
		return v
	}
	return -1
}

// ------------------------------------------------------------

// LogInterceptor defines an object which is able to
// process individual LogRecord instances
type LogInterceptor interface {
	ProcItem(appType string, record *LogRecord)
}

// ------------------------------------------------------------

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

// NewParser creates a new instance of the Parser.
// localTimezone has format: "(-|+)[0-9]{2}:[0-9]{2}"
func NewParser(path string, localTimezone string) *Parser {
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

// Parser parses a single file represented by fr Scanner.
// Because KonText does not log (at least currently) a timezone info,
// this information is also required to process the log properly.
type Parser struct {
	fr            *bufio.Scanner
	localTimezone string
}

func (p *Parser) parseLine(s string) (LogRecord, error) {
	jsonLine := parseRawLine(s)
	var record LogRecord
	var err error
	if jsonLine != "" {
		err = json.Unmarshal([]byte(jsonLine), &record)
		if err == nil {
			record.Date = importDatetimeString(record.Date, p.localTimezone)
		}

	} else if tp := getLineType(s); tp == "QUERY" {
		err = fmt.Errorf("Failed to process QUERY entry: %s", s)
	}
	return record, err
}

// Parse runs the parsing process based on provided minimum accepted record
// time, record type (which is just passed to ElastiSearch) and a
// provided LogInterceptor).
func (p *Parser) Parse(fromTimestamp int64, recType string, proc LogInterceptor) {
	for p.fr.Scan() {
		rec, err := p.parseLine(p.fr.Text())
		if err == nil {
			if rec.GetTime().Unix() >= fromTimestamp {
				proc.ProcItem(recType, &rec)
			}

		} else {
			log.Printf("Error parsing record: %s", err)
		}
	}
}
