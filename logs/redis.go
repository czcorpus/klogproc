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
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

// GetRecordsInRedis loads log data from a Redis queue (list type).
// The data is expected to be in JSON format.
func GetRecordsInRedis(address string, database int, queueKey string) ([]LogRecord, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       database,
	})

	if queueKey == "" {
		return nil, fmt.Errorf("No queue key provided")
	}

	size := int(client.LLen(queueKey).Val())
	ans := make([]LogRecord, size)

	for i := 0; i < size; i++ {
		rawItem, err := client.LPop(queueKey).Bytes()
		if err != nil {
			return nil, err
		}
		var item LogRecord
		err = json.Unmarshal(rawItem, &item)
		if err != nil {
			return nil, err
		}
		ans[i] = item
	}

	return ans, nil
}
