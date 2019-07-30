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

package celery

import "fmt"

// for more info on the values below, see
// https://docs.celeryproject.org/en/latest/userguide/workers.html#statistics

type Rusage struct {
	IDRss    int     `json:"idrss"`
	Inblock  int     `json:"inblock"`
	IsRss    int     `json:"isrss"`
	IxRss    int     `json:"ixrss"`
	Majflt   int     `json:"majflt"`
	MaxRss   int     `json:"maxrss"`
	Minflt   int     `json:"minflt"`
	MsgRcv   int     `json:"msgrcv"`
	MsgSnd   int     `json:"msgsnd"`
	Nivcsw   int     `json:"nivcsw"`
	Nsignals int     `json:"nsignals"`
	Nswap    int     `json:"nswap"`
	Nvcsw    int     `json:"nvcsw"`
	Oublock  int     `json:"oublock"`
	Stime    float32 `json:"stime"`
	Utime    float32 `json:"utime"`
}

type Broker struct {
	Alternates       []string               `json:"alternates"` // TODO is the type right?
	ConnectTimeout   int                    `json:"connect_timeout"`
	FailoverStrategy string                 `json:"failover_strategy"`
	Heartbeat        float32                `json:"heartbeat"`
	Hostname         string                 `json:"hostname"`
	Insist           bool                   `json:"insist"`
	LoginMethod      string                 `json:"login_method"` // TODO is type right?
	Port             int                    `json:"port"`
	SSL              bool                   `json:"ssl"`
	Transport        string                 `json:"transport"`
	TransportOptions map[string]interface{} `json:"transport_options"` // TODO type
	URIPrefix        string                 `json:"uri_prefix"`
	UserID           int                    `json:"userid"` // TODO type
	VirtualHost      string                 `json:"virtual_host"`
}

type PoolWrites struct {
	All      string `json:"all"`
	Avg      string `json:"avg"`
	Inqueues struct {
		Active int `json:"active"`
		Total  int `json:"total"`
	} `json:"inqueues"`
	Raw      string `json:"raw"`
	Strategy string `json:"strategy"`
	Total    int    `json:"total"`
}

type Pool struct {
	MaxConcurrency        int        `json:"max-concurrency"`
	MaxTasksPerChild      string     `json:"max-tasks-per-child"` // TODO maybe a combined type "N/A" vs integers
	Processes             []int      `json:"processes"`
	PutGuardedBySemaphore bool       `json:"put-guarded-by-semaphore"`
	Timeouts              []float32  `json:"timeouts"`
	Writes                PoolWrites `json:"writes"`
}

type InputRecord struct {
	Broker        Broker         `json:"broker"`
	Clock         string         `json:"clock"`
	PID           int            `json:"pid"`
	Pool          Pool           `json:"pool"`
	PrefetchCount int            `json:"prefetch_count"`
	Rusage        Rusage         `json:"rusage"`
	Total         map[string]int `json:"total"`
}

func (r InputRecord) String() string {
	return fmt.Sprintf("InputRecord{Broker: %v, Clock: %v, PID: %d, Pool: %v, PrefetchCount: %d, Rusage: %v, Total: %v}",
		r.Broker, r.Clock, r.PID, r.Pool, r.PrefetchCount, r.Rusage, r.Total)
}
