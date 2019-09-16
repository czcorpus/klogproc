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

package ske

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/fsop"
)

// loadUserMap loads json-encoded [username]=>[user_id] map
func loadUserMap(path string) (map[string]int, error) {
	fr, err := os.OpenFile(path, os.O_RDONLY, 0644)
	byteValue, err := ioutil.ReadAll(fr)
	if err != nil {
		return nil, fmt.Errorf("Failed to load usermap.json file: %s", err)
	}
	ans := make(map[string]int)
	err = json.Unmarshal(byteValue, &ans)
	if err != nil {
		err = fmt.Errorf("Failed to parse usermap.json file: %s", err)
	}
	return ans, err
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	userMap map[string]int
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, anonymousUsers []int) (*OutputRecord, error) {

	userID := -1
	if logRecord.User != "-" && logRecord.User != "" {
		uid, ok := t.userMap[logRecord.User]
		if !ok {
			return nil, fmt.Errorf("Failed to find user ID of [%s]", logRecord.User)
		}
		userID = uid
	}

	corpname, isLimited := importCorpname(logRecord.Corpus)
	r := &OutputRecord{
		Type:        recType,
		time:        logRecord.GetTime(),
		Datetime:    logRecord.GetTime().Format(time.RFC3339),
		IPAddress:   logRecord.Request.RemoteAddr,
		UserAgent:   logRecord.Request.HTTPUserAgent,
		IsAnonymous: userID == -1 || conversion.UserBelongsToList(userID, anonymousUsers),
		IsQuery:     isEntryQuery(logRecord.Action),
		UserID:      strconv.Itoa(userID),
		Action:      logRecord.Action,
		Corpus:      corpname,
		Limited:     isLimited,
		Subcorpus:   logRecord.Subcorpus,
		ProcTime:    logRecord.ProcTime,
	}
	r.ID = createID(r)
	return r, nil
}

// NewTransformer is a default constructor for the Transformer.
// It also loads user ID map from a configured file (if exists).
func NewTransformer(customConfDir string) (*Transformer, error) {
	ans := &Transformer{
		userMap: make(map[string]int),
	}
	var err error
	if customConfDir != "" {
		confPath := filepath.Join(customConfDir, "usermap.json")
		if fsop.IsFile(confPath) {
			log.Printf("INFO: loading custom user map %s", confPath)
			userMap, err := loadUserMap(confPath)
			if err != nil {
				return ans, err
			}
			ans.userMap = userMap
		}
	}
	return ans, err
}
