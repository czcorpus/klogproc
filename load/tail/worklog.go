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
	"io"
	"os"

	"klogproc/fsop"
	"klogproc/servicelog"

	"github.com/czcorpus/cnc-gokit/collections"

	"github.com/rs/zerolog/log"
)

type updateRequest struct {
	FilePath string
	Value    servicelog.LogRange
}

// WorklogRecord provides log reading position info for all configured apps
type WorklogRecord = map[string]servicelog.LogRange

// Worklog provides functions to store/retrieve information about
// file reading operations to be able to continue in case of an
// interruption/error. Worklog can handle incoming status updates
// even if they arrive out of order - which is rather a typical
// situation (e.g. ignored lines are confirmed sooner that the ones
// send to Elastic/Influx).
type Worklog struct {
	filePath    string
	fr          *os.File
	rec         *collections.ConcurrentMap[string, servicelog.LogRange]
	updRequests chan updateRequest
}

// Init initializes the worklog. It must be called before any other
// operation.
func (w *Worklog) Init() error {
	var err error
	if w.filePath == "" {
		return fmt.Errorf("failed to initialize tail worklog - no path specified")
	}
	log.Info().Msgf("Initializing worklog %s", w.filePath)
	w.fr, err = os.OpenFile(w.filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	byteValue, err := io.ReadAll(w.fr)
	if err != nil {
		return err
	}
	if len(byteValue) > 0 {
		var err error
		w.rec, err = collections.NewConcurrentMapFromJSON[string, servicelog.LogRange](byteValue)
		if err != nil {
			return err
		}
	}
	w.updRequests = make(chan updateRequest)
	go func() {
		for req := range w.updRequests {
			curr := w.rec.Get(req.FilePath)
			if curr.Inode != req.Value.Inode {
				log.Warn().Msgf("inode for %s has changed from %d to %d", req.FilePath, curr.Inode, req.Value.Inode)
			}
			// rules for worklog update:
			// 1) if inodes differ then write the new record
			// 2) non-written incoming item always overwrites a written one (to make sure we try again from its position)
			// 3) non-written incoming rewrites the current written no matter how old it is
			// 4) written incoming item can fix current non-written if its older or of the same age
			// 5) if both are written then only more recent (higher seek) can overwrite the current one
			if curr.Inode != req.Value.Inode ||
				!curr.Written && curr.SeekStart >= req.Value.SeekStart ||
				curr.Written && req.Value.SeekEnd >= curr.SeekEnd ||
				!req.Value.Written && (curr.Written || req.Value.SeekEnd < curr.SeekEnd) {
				w.rec.Set(req.FilePath, req.Value)
				w.save()

			} else {
				log.Warn().Msgf("worklog[%s] item %v won't be saved due to the current %v", req.FilePath, req.Value, curr)
			}
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
func (w *Worklog) UpdateFileInfo(filePath string, logPosition servicelog.LogRange) {
	w.updRequests <- updateRequest{
		FilePath: filePath,
		Value:    logPosition,
	}
}

// ResetFile sets a zero seek and line for a new or an existing file.
// Returns an inode of a respective file and a possible error
func (w *Worklog) ResetFile(filePath string) (int64, error) {
	inode, _, err := fsop.GetFileProps(filePath)
	if err != nil {
		return -1, err
	}
	w.updRequests <- updateRequest{
		FilePath: filePath,
		Value: servicelog.LogRange{
			Inode:     inode,
			SeekStart: 0,
			SeekEnd:   0,
			Written:   true, // otherwise update request won't be accepted
		},
	}
	return inode, nil
}

// GetData retrieves reading info for a provided app
func (w *Worklog) GetData(filePath string) servicelog.LogRange {
	v, ok := w.rec.GetWithTest(filePath)
	if ok {
		return v
	}
	return servicelog.LogRange{Inode: -1, SeekStart: 0, SeekEnd: 0}
}

// NewWorklog creates a new Worklog instance. Please note that
// Init() must be called before you can begin using the worklog.
func NewWorklog(path string) *Worklog {
	return &Worklog{
		filePath: path,
		rec:      collections.NewConcurrentMap[string, servicelog.LogRange](),
	}
}
