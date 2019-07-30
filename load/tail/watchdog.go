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
	"sync"
	"syscall"
	"time"
)

const (
	defaultTickerIntervalSecs = 60
)

// FileConf represents a configuration for a single
// log file to be watched
type FileConf struct {
	Path    string `json:"path"`
	AppType string `json:"appType"`
	Version int    `json:"version"`
}

// Conf wraps all the configuration for the 'tail' function
type Conf struct {
	IntervalSecs int        `json:"intervalSecs"`
	WorklogPath  string     `json:"worklogPath"`
	Files        []FileConf `json:"files"`
}

// FileTailProcessor specifies an object which is able to utilize all
// the "events" watchdog provides when processing a file tail for
// a concrete appType
type FileTailProcessor interface {
	AppType() string
	FilePath() string
	CheckIntervalSecs() int
	OnCheckStart()
	OnEntry(item string)
	OnCheckStop()
	OnQuit()
}

// Run starts the process of (multiple) log watching
func Run(conf *Conf, processors []FileTailProcessor, finishEvent chan<- bool) {
	tickerInterval := time.Duration(conf.IntervalSecs)
	if tickerInterval == 0 {
		log.Printf("WARNING: intervalSecs for tail mode not set, using default %ds", defaultTickerIntervalSecs)
		tickerInterval = time.Duration(defaultTickerIntervalSecs)

	} else {
		log.Printf("INFO: configured to check for file changes every %d second(s)", tickerInterval)
	}
	ticker := time.NewTicker(tickerInterval * time.Second)
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
		readers = make([]*FileTailReader, len(processors))
		for i, processor := range processors {
			wlItem := worklog.GetData(processor.AppType())
			log.Printf("INFO: Found configuration for file %s", processor.FilePath())
			if wlItem.Inode > -1 {
				log.Printf("INFO: Found worklog for %s, inode: %d, seek: %d", processor.FilePath(), wlItem.Inode, wlItem.Seek)
			}
			rdr, err := NewReader(processor, wlItem.Inode, wlItem.Seek)
			if err != nil {
				log.Print("ERROR: ", err)
				quitChan <- true
			}
			readers[i] = rdr
		}
	}

	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup
			wg.Add(len(readers))
			for _, reader := range readers {
				go func(rdr *FileTailReader) {
					rdr.Processor().OnCheckStart()
					rdr.ApplyNewContent(
						func(v string) {
							rdr.Processor().OnEntry(v)
						},
						func(inode int64, seek int64) {
							worklog.UpdateFileInfo(rdr.AppType(), inode, seek)
						},
					)
					rdr.Processor().OnCheckStop()
					wg.Done()
				}(reader)
			}
			wg.Wait()
		case quit := <-quitChan:
			if quit {
				ticker.Stop()
				for _, processor := range processors {
					processor.OnQuit()
				}
				worklog.Close()
				finishEvent <- true
			}
		case <-syscallChan:
			log.Print("INFO: Caught signal, exiting...")
			ticker.Stop()
			for _, reader := range readers {
				reader.Processor().OnQuit()
			}
			worklog.Close()
			finishEvent <- true
		}
	}
}
