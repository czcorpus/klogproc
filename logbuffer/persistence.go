// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package logbuffer

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
)

func (st *Storage[T, U]) mkStorageFileName() string {
	h := fnv.New32a()
	h.Write([]byte(st.logFilePath))
	return fmt.Sprintf("%x.json", h.Sum32())
}

func (st *Storage[T, U]) SetStateData(stateData U) {
	st.stateData = stateData
	st.stateWriting <- stateData
}

func (st *Storage[T, U]) LoadStateData() (U, error) {
	var ans U
	fullPath := filepath.Join(st.storageDirPath, st.mkStorageFileName())
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ans, fmt.Errorf("failed to read log buffer state data: %w", err)
	}
	err = json.Unmarshal(data, &ans)
	if err != nil {
		return ans, fmt.Errorf("failed to unmarshal log buffer state data: %w", err)
	}
	return ans, nil
}

func (st *Storage[T, U]) GetStateData() U {
	return st.stateData
}
