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

package redisdb

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type ActivityReport struct {
	NumRequests    int     `json:"numRequests"`
	TimeWindowSecs int     `json:"timeWindowSecs"`
	Mean           float64 `json:"mean"`
	Stdev          float64 `json:"stdev"`
	Created        string  `json:"created"`
	TotalReports   int     `json:"totalReports"`
}

// ConnectionConf specifies a configuration required to store data
// to a Redis database
type ConnectionConf struct {
	// Address contains both host and port (e.g. localhost:6379)
	Address        string `json:"address"`
	DB             int    `json:"db"`
	IPBlockListKey string `json:"ipBlockListKey"`
}

type RedisWriter struct {
	client       *redis.Client
	ctx          context.Context
	blockListKey string
}

func (rw *RedisWriter) WriteReport(ip string, report *ActivityReport) error {
	if rw.client == nil {
		return nil
	}
	cmd := rw.client.HGet(rw.ctx, rw.blockListKey, ip)
	if cmd.Err() == redis.Nil {
		report.TotalReports = 1

	} else if cmd.Err() != nil {
		return cmd.Err()

	} else {
		var prevReport ActivityReport
		err := json.Unmarshal([]byte(cmd.Val()), &prevReport)
		if err != nil {
			return err
		}
		report.TotalReports = prevReport.TotalReports + 1
	}

	jsn, err := json.Marshal(report)
	if err != nil {
		return err
	}
	cmd2 := rw.client.HSet(rw.ctx, rw.blockListKey, ip, jsn)
	if cmd2.Err() != nil {
		return cmd2.Err()
	}
	return nil
}

func NewRedisWriter(conf ConnectionConf) *RedisWriter {
	var client *redis.Client
	if conf.Address != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     conf.Address,
			Password: "",
			DB:       conf.DB,
		})
	}
	return &RedisWriter{
		client:       client,
		ctx:          context.Background(),
		blockListKey: conf.IPBlockListKey,
	}
}
