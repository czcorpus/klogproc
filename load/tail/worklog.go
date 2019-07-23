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

package tail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// WorklogItem stores inode & seek position of last read operation
type WorklogItem struct {
	Inode int64 `json:"inode"`
	Seek  int64 `json:"seek"`
}

type updateRequest struct {
	AppType string
	Value   WorklogItem
}

// WorklogRecord provides WorkLogItem info for all configured apps
type WorklogRecord = map[string]WorklogItem

// Worklog provides functions to store/retrieve information about
// file reading operations to be able to continue in case of an
// interruption
type Worklog struct {
	filePath    string
	fr          *os.File
	rec         WorklogRecord
	updRequests chan updateRequest
}

// Init initializes the worklog. It must be called before any other
// operation.
func (w *Worklog) Init() error {
	var err error
	if w.filePath == "" {
		return fmt.Errorf("Failed to initialize tail worklog - no path specified")
	}
	log.Printf("Initializing worklog %s", w.filePath)
	w.fr, err = os.OpenFile(w.filePath, os.O_CREATE|os.O_RDWR, 0644)
	byteValue, err := ioutil.ReadAll(w.fr)
	if err != nil {
		return err
	}
	if len(byteValue) > 0 {
		err := json.Unmarshal(byteValue, &w.rec)
		if err != nil {
			return err
		}
	}
	w.updRequests = make(chan updateRequest)
	go func() {
		for req := range w.updRequests {
			w.rec[req.AppType] = req.Value
			w.save()
		}
	}()
	return nil
}

// Close cleans up worklog for safe exit
func (w *Worklog) Close() {
	if w.fr != nil {
		w.fr.Close()
	}
	if w.updRequests != nil {
		close(w.updRequests)
	}
}

// save stores worklog's state to a configured file.
// It is called automatically after each log update
// request is processed.
func (w *Worklog) save() error {
	err := w.fr.Truncate(0)
	if err != nil {
		return err
	}
	_, err = w.fr.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}
	data, err := json.Marshal(w.rec)
	if err != nil {
		return err
	}
	_, err = w.fr.Write(data)
	if err != nil {
		return err
	}
	err = w.fr.Sync()
	if err != nil {
		return err
	}
	return nil
}

// UpdateFileInfo adds individual app reading position info. Please
// note that this does not save the worklog.
func (w *Worklog) UpdateFileInfo(appType string, inode int64, seek int64) {
	w.updRequests <- updateRequest{AppType: appType, Value: WorklogItem{Inode: inode, Seek: seek}}
}

// GetData retrieves reading info for a provided app
func (w *Worklog) GetData(appType string) WorklogItem {
	v, ok := w.rec[appType]
	if ok {
		return v
	}
	return WorklogItem{Inode: -1, Seek: 0}
}

// NewWorklog creates a new Worklog instance. Please note that
// Init() must be called before you can begin using the worklog.
func NewWorklog(path string) *Worklog {
	return &Worklog{filePath: path, rec: make(map[string]WorklogItem)}
}
