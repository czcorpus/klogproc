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

package sredis

import (
	"fmt"
	"log"

	"github.com/czcorpus/klogproc/transform"
	"github.com/czcorpus/klogproc/transform/kontext"
	"github.com/go-redis/redis"
)

// RedisConf is a structure containing information
// about Redis database containing logs to be
// processed.
type RedisConf struct {
	Address  string `json:"address"`
	Database int    `json:"database"`
	QueueKey string `json:"queueKey"`
}

// RedisQueue provides access to Redis database containing
// KonText log records.
type RedisQueue struct {
	db             *redis.Client
	queueKey       string
	failedItemsKey string
	localTimezone  string
}

// OpenRedisQueue creates a client for Redis
func OpenRedisQueue(address string, database int, queueKey string, localTimezone string) (*RedisQueue, error) {
	if queueKey == "" {
		return nil, fmt.Errorf("No queue key provided")
	}
	client := &RedisQueue{
		db: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: "",
			DB:       database,
		}),
		queueKey:       queueKey,
		failedItemsKey: queueKey + "_failed",
		localTimezone:  localTimezone,
	}
	return client, nil
}

// GetItems loads log data from a Redis queue (list type).
// The data is expected to be in JSON format.
//
// Please note that invalid records are taken from queue too
// and then thrown away (with logged message containing the
// original item source).
func (rc *RedisQueue) GetItems() []transform.InputRecord {

	size := int(rc.db.LLen(rc.queueKey).Val())
	log.Printf("INFO: Found %d records in log queue", size)
	ans := make([]transform.InputRecord, 0, size)

	for i := 0; i < size; i++ {
		rawItem, err := rc.db.LPop(rc.queueKey).Bytes()
		if err != nil {
			log.Printf("WARNING: %s, orig item: %s", err, rawItem)
		}
		item, err := kontext.ImportJSONLog(rawItem, rc.localTimezone)
		if err != nil {
			log.Printf("WARNING: %s, orig item: %s", err, rawItem)

		} else {
			ans = append(ans, item)
		}
	}
	return ans
}

// RescueFailedChunks puts records to the end of the Redis queue.
// This is mostly for handling ElasticSearch import errors.
func (rc *RedisQueue) RescueFailedChunks(data [][]byte) error {
	for _, item := range data {
		rc.db.RPush(rc.failedItemsKey, item)
	}
	if len(data) > 0 {
		log.Printf("INFO: Stored raw data to be reinserted next time (num bulk insert rows: %d)", len(data))
	}
	return nil
}

// GetRescuedChunksIterator returns an iterator object
// for rescued raw bulk insert records.
func (rc *RedisQueue) GetRescuedChunksIterator() *RedisRescuedChunkIterator {
	return &RedisRescuedChunkIterator{
		db:      rc.db,
		currPos: 0,
		dbKey:   rc.failedItemsKey,
	}
}

// ------

// RedisRescuedChunkIterator provides stateful access to
// individual bulk insert chunks ([meta line, data line]+  "new line")
type RedisRescuedChunkIterator struct {
	db      *redis.Client
	currPos int64
	dbKey   string
}

// GetNextChunk provide next chunk of bulk insert data.
// If nothing is found then a slice of size 0 is returned.
func (rrci *RedisRescuedChunkIterator) GetNextChunk() [][]byte {
	queueSize := rrci.db.LLen(rrci.dbKey).Val()
	tmp := make([][]byte, 0, queueSize)
	var curr string
	for ; rrci.currPos < queueSize && curr != "\n"; rrci.currPos++ {
		currSrch := rrci.db.LRange(rrci.dbKey, rrci.currPos, rrci.currPos).Val()
		if len(currSrch) == 1 {
			tmp = append(tmp, []byte(currSrch[0]))
		}
	}
	return tmp
}

// RemoveVisitedItems removes from Redis all the items we iterated through so far
func (rrci *RedisRescuedChunkIterator) RemoveVisitedItems() (int, error) {
	status := rrci.db.LTrim(rrci.dbKey, rrci.currPos, int64(-1))
	if status.Err() != nil {
		return 0, status.Err()
	}
	numProc := int(rrci.currPos)
	rrci.currPos = 0
	return numProc, nil
}
