// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package scripting

import (
	"klogproc/load"
	"klogproc/servicelog"
	"net"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type dummyInputRec struct {
	ID                  string
	Time                time.Time
	ClientIP            net.IP
	UserAgent           string
	_ClusteringClientID string
	_ClusterSize        int
	Args                struct {
		Name     string
		Position int
	}
}

func (r *dummyInputRec) GetTime() time.Time {
	return r.Time
}

func (r *dummyInputRec) GetClientIP() net.IP {
	return r.ClientIP
}

func (r *dummyInputRec) GetUserAgent() string {
	return r.UserAgent
}

func (r *dummyInputRec) ClusteringClientID() string {
	return r._ClusteringClientID
}

func (r *dummyInputRec) ClusterSize() int {
	return r._ClusterSize
}

func (r *dummyInputRec) SetCluster(size int) {
	r._ClusterSize = size
}

func (r *dummyInputRec) IsProcessable() bool {
	return true
}

func (r *dummyInputRec) IsSuspicious() bool {
	return false
}

// -------

type dummyOutRec struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	time        time.Time
	Time        string `json:"string"`
	IsAI        bool   `json:"isAi"`
	CustomField string `json:"customField"`
}

func (r *dummyOutRec) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
	// NOP
}

// ToJSON creates an object suitable for storing to ElasticSearch, CouchDB and other
// document-oriented databases
func (r *dummyOutRec) ToJSON() ([]byte, error) {
	return []byte{}, nil
}

// Create an idempotent unique identifier of the record.
// This can be typically acomplished by hashing the original
// log record.
func (r *dummyOutRec) GetID() string {
	return r.ID
}

// Return app type as defined by an external convention
// (e.g. for UCNK: kontext, syd, morfio, treq,...)
func (r *dummyOutRec) GetType() string {
	return r.Type
}

// Get time of the log record
func (r *dummyOutRec) GetTime() time.Time {
	return r.time
}

func (r *dummyOutRec) SetTime(t time.Time) {
	r.Time = t.Format(time.RFC3339)
	r.time = t
}

func (r *dummyOutRec) LSetProperty(name string, value lua.LValue) error {
	return ErrScriptingNotSupported
}

func (r *dummyOutRec) GenerateDeterministicID() string {
	return "dummy-out-rec-01"
}

// --------

type ltrans struct {
}

func (lt *ltrans) AppType() string {
	return "dummy"
}

func (lt *ltrans) HistoryLookupItems() int { return 0 }

func (lt *ltrans) Preprocess(
	rec servicelog.InputRecord,
	prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	return []servicelog.InputRecord{rec}, nil
}

func (lt *ltrans) Transform(
	logRec servicelog.InputRecord,
) (servicelog.OutputRecord, error) {
	tLogRec, ok := logRec.(*dummyInputRec)
	if !ok {
		panic("invalid type")
	}
	return &dummyOutRec{
		ID:          tLogRec.ID,
		Time:        tLogRec.Time.Format("2006-01-02T15:04:05-07:00"),
		IsAI:        false,
		CustomField: "custom data for " + tLogRec.ID,
	}, nil
}

// ------------------

type dummyConf struct {
	AppType        string
	Version        string
	AnonymousUsers []int
	ScriptPath     string
}

func (c *dummyConf) GetAppType() string {
	return c.AppType
}

func (c *dummyConf) GetVersion() string {
	return c.Version
}

func (c *dummyConf) GetBuffer() *load.BufferConf {
	return &load.BufferConf{}
}

func (c *dummyConf) GetScriptPath() string {
	return c.ScriptPath
}

func (c *dummyConf) ClusterSize() int {
	return 1
}
