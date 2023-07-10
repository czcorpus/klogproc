// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

package kontext013

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"klogproc/conversion"
)

var (
	datetimeRegexp = regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2})(\\s|T)([012]\\d:[0-5]\\d:[0-5]\\d(\\.\\d+))")
)

func importDatetimeString(dateStr string) (string, error) {
	srch := datetimeRegexp.FindStringSubmatch(dateStr)
	if len(srch) > 0 {
		return fmt.Sprintf("%sT%s", srch[1], srch[3]), nil
	}
	return "", fmt.Errorf("Failed to import datetime \"%s\"", dateStr)
}

// ImportJSONLog parses original JSON record with some
// additional value corrections.
func ImportJSONLog(jsonLine []byte) (*InputRecord, error) {
	var record InputRecord
	err := json.Unmarshal(jsonLine, &record)
	if err != nil {
		return nil, err
	}
	dt, err := importDatetimeString(record.Date)
	if err != nil {
		return nil, err
	}
	record.Date = dt
	return &record, nil
}

// ------------------------------------------------------------

// Request is a simple representation of
// HTTP request metadata used in KonText logging
type Request struct {
	HTTPForwardedFor string `json:"HTTP_X_FORWARDED_FOR"`
	HTTPUserAgent    string `json:"HTTP_USER_AGENT"`
	HTTPRemoteAddr   string `json:"HTTP_REMOTE_ADDR"`
	RemoteAddr       string `json:"REMOTE_ADDR"`
}

// ------------------------------------------------------------

// ErrorRecord specifies a thrown error along with
// optional anchor for easier search within text file
// log
type ErrorRecord struct {
	Name   string `json:"name"`
	Anchor string `json:"anchor"`
}

// ------------------------------------------------------------

// InputRecord represents a parsed KonText record
type InputRecord struct {
	UserID   int                    `json:"user_id"`
	ProcTime float32                `json:"proc_time"`
	Date     string                 `json:"date"`
	Action   string                 `json:"action"`
	Request  Request                `json:"request"`
	Params   map[string]interface{} `json:"params"`
	PID      int                    `json:"pid"`
	Settings map[string]interface{} `json:"settings"`
	Error    ErrorRecord            `json:"error"`
}

// GetTime returns record's time as a Golang's Time
// instance. Please note that the value is truncated
// to seconds.
func (rec *InputRecord) GetTime() time.Time {
	return conversion.ConvertDatetimeStringWithMillisNoTZ(rec.Date)
}

// GetClientIP returns a client IP no matter in which
// part of the record it was found
// (e.g. REMOTE_ADDR vs. HTTP_REMOTE_ADDR vs. HTTP_FORWARDED_FOR)
func (rec *InputRecord) GetClientIP() net.IP {
	if rec.Request.HTTPForwardedFor != "" {
		return net.ParseIP(rec.Request.HTTPForwardedFor)

	} else if rec.Request.HTTPRemoteAddr != "" {
		return net.ParseIP(rec.Request.HTTPRemoteAddr)

	} else if rec.Request.RemoteAddr != "" {
		return net.ParseIP(rec.Request.RemoteAddr)
	}
	return make([]byte, 0)
}

func (rec *InputRecord) ClusteringClientID() string {
	return conversion.GenerateRandomClusteringID()
}

// GetUserAgent returns a raw HTTP user agent info as provided by the client
func (rec *InputRecord) GetUserAgent() string {
	return rec.Request.HTTPUserAgent
}

// IsProcessable returns true if there was no error in reading the record
func (rec *InputRecord) IsProcessable() bool {
	return true
}

// GetStringParam fetches a string parameter from
// a special "params" sub-object
func (rec *InputRecord) GetStringParam(name string) string {
	switch v := rec.Params[name].(type) {
	case string:
		return v
	}
	return ""
}

// GetIntParam fetches an integer parameter from
// a special "params" sub-object
func (rec *InputRecord) GetIntParam(name string) int {
	switch v := rec.Params[name].(type) {
	case int:
		return v
	}
	return -1
}

// GetAlignedCorpora fetches aligned corpora names from arguments
// found in record's "Params" attribute. It isolates
// user from miscellaneous idiosyncrasies of KonText/Bonito
// URL parameter handling (= it's not always that straightforward
// to detect aligned languages from raw URL).
func (rec *InputRecord) GetAlignedCorpora() []string {
	tmp := make(map[string]bool)
	for k := range rec.Params {
		if strings.HasPrefix(k, "queryselector_") {
			tmp[k[len("queryselector_"):]] = true
		}
		if strings.HasPrefix(k, "pcq_pos_neg_") {
			tmp[k[len("pcq_pos_neg_"):]] = true
		}
	}
	ans := make([]string, len(tmp))
	i := 0
	for k := range tmp {
		ans[i] = k
		i++
	}
	return ans
}
