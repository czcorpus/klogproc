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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// FileConf represents a configuration for a single
// log file to be watched
type FileConf struct {
	Path    string `json:"path"`
	AppType string `json:"appType"`
}

// Conf wraps all the configuration for the 'tail' function
type Conf struct {
	IntervalSecs int        `json:"intervalSecs"`
	WorklogPath  string     `json:"worklogPath"`
	Files        []FileConf `json:"files"`
}

// Run starts the process of (multiple) log watching
func Run(conf *Conf, onEntry func(item string, appType string), onStop func()) {
	ticker := time.NewTicker(5 * time.Second)
	quitChan := make(chan bool, 10)
	syscallChan := make(chan os.Signal, 10)
	signal.Notify(syscallChan, os.Interrupt)
	signal.Notify(syscallChan, syscall.SIGTERM)
	worklog := NewWorklog(conf.WorklogPath)
	var readers []*FileTailReader

	err := worklog.Init()
	if err != nil {
		log.Print("ERROR: ", err)
		quitChan <- true

	} else {
		readers = make([]*FileTailReader, 0, 50)
		for _, fc := range conf.Files {
			wlItem := worklog.GetData(fc.AppType)
			log.Printf("INFO: Found configuration for file %s", fc.Path)
			if wlItem.Inode > -1 {
				log.Printf("INFO: Found worklog for %s, inode: %d, seek: %d", fc.Path, wlItem.Inode, wlItem.Seek)
			}
			rdr, err := NewReader(fc.Path, fc.AppType, wlItem.Inode, wlItem.Seek)
			if err != nil {
				log.Print("ERROR: ", err)
				quitChan <- true
			}
			readers = append(readers, rdr)
		}
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, f := range conf.Files {
					for _, reader := range readers {
						reader.ApplyNewContent(
							func(v string) {
								onEntry(v, f.AppType)
							},
							func(inode int64, seek int64) {
								worklog.UpdateFileInfo(f.AppType, inode, seek)
								worklog.Save()
							},
						)
					}
				}
			case quit := <-quitChan:
				if quit {
					ticker.Stop()
					onStop()
					worklog.Close()
				}
			case <-syscallChan:
				log.Print("INFO: Caught signal, exiting...")
				ticker.Stop()
				onStop()
				worklog.Close()
			}
		}
	}()
}
