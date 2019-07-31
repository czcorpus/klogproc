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

import (
	"encoding/json"
	"fmt"
	"time"
)

// OutputRecord contains a set of values from Celery inspect
// transformed to the form usable for us
type OutputRecord struct {
	ID                string         `json:"id"`
	Clock             int            `json:"clock"`
	Hostname          string         `json:"hostname"`
	ProcTime          float32        `json:"procTime"` // ProcTime = utime + stime
	NumWorkerRestarts int            `json:"numWorkerRestart"`
	NumTaskCalls      map[string]int `json:"numTaskCalls"`
	NumTasksTotal     int            `json:"numTasksTotal"`
}

// ToJSON converts data to a JSON document (typically for ElasticSearch)
func (r *OutputRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToInfluxDB creates tags and values to store in InfluxDB
func (r *OutputRecord) ToInfluxDB() (tags map[string]string, values map[string]interface{}) {
	tags = make(map[string]string)
	tags["host"] = r.Hostname

	values = make(map[string]interface{})
	for k, v := range r.NumTaskCalls {
		values[fmt.Sprintf("task_%s", k)] = v
	}
	values["cpu_time"] = r.ProcTime
	values["num_worker_restarts"] = r.NumWorkerRestarts
	values["num_tasks_total"] = r.NumTasksTotal
	values["clock"] = r.Clock
	return
}

func (cnkr *OutputRecord) GetID() string {
	return cnkr.ID
}

func (cnkr *OutputRecord) GetType() string {
	return "celery"
}

// GetTime returns Go Time instance representing
// date and time when the record was created.
func (cnkr *OutputRecord) GetTime() time.Time {
	return time.Now()
}

func (cnkr *OutputRecord) SetLocation(countryName string, latitude float32, longitude float32, timezone string) {
}
