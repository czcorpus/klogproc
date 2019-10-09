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

package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type UserMap struct {
	data          map[string]int
	ignoreMissing bool
}

func (um *UserMap) GetIdOf(username string) int {
	v, ok := um.data[username]
	if ok || um.ignoreMissing {
		return v
	}
	return -1
}

// LoadUserMap loads json-encoded [username]=>[user_id] map
func LoadUserMap(path string) (*UserMap, error) {
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
	return &UserMap{data: ans, ignoreMissing: false}, err
}

func EmptyUserMap() *UserMap {
	return &UserMap{data: make(map[string]int), ignoreMissing: true}
}
