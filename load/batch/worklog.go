// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
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

// fileselect functions are used to find proper KonText application log files
// based on logs processed so far. Please note that in recent KonText and
// Klogproc versions this is rather a fallback/offline functionality.

package batch

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func worklogExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Worklog represents an object keeping track
// of the last time the Save() has been called.
// It is used to find new log files which have
// not been yet processed.
type Worklog struct {
	filePath string
}

func (w *Worklog) String() string {
	return fmt.Sprintf("Worklog{filePath: %s}", w.filePath)
}

// GetLastRecord returns a last UNIX timestamp
// generated by the Save() method.
//
// In case no real log exists in filesystem,
// -1 is returned.
func (w *Worklog) GetLastRecord() int64 {
	if worklogExists(w.filePath) {
		f, err := os.Open(w.filePath)
		if err != nil {
			panic(fmt.Sprintf("Failed to open existing worklog file: %s", err))
		}
		defer f.Close()
		reader := bufio.NewScanner(f)
		lastLine := ""
		if reader != nil {
			for reader.Scan() {
				tmp := strings.TrimSpace(reader.Text())
				if tmp[0] != '#' {
					lastLine = tmp
				}
			}
		}
		if lastLine != "" {
			ans, err := strconv.ParseInt(lastLine, 10, 64)
			if err == nil {
				return ans
			}
		}

	} else {
		log.Info().Msg("No worklog file present - all the found log files will be processed")
	}
	return -1
}

// Save saves the worklog by writing actual UNIX timestamp
// to the log file.
func (w *Worklog) Save() error {
	f, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to open worklog file for writing: %s", err))
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	defer writer.Flush()
	_, err = writer.WriteString(fmt.Sprintf("%d\n", time.Now().Unix()))
	return err
}

func (w *Worklog) RescueFailedChunks(data [][]byte) error {
	// TODO we do nothing here but we should move
	// status pointer before this broken chunk
	return nil
}

func (w *Worklog) Reset() error {
	return os.Truncate(w.filePath, 0)
}

// NewWorklog creates an instance of Worklog with
// defined path. No file access attempts are made.
func NewWorklog(path string) *Worklog {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			log.Fatal().Msgf("cannot create worklog - %s", err)
		}
		f.Close()
	}
	return &Worklog{filePath: path}
}
