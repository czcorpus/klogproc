// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

type FileConf struct {
	Path    string `json:"path"`
	AppType string `json:"appType"`
}

type Conf struct {
	IntervalSecs int        `json:"intervalSecs"`
	Files        []FileConf `json:"files"`
}

func Run(conf *Conf, onEntry func(item string, appType string), onStop func()) {
	ticker := time.NewTicker(5 * time.Second)
	quitChan := make(chan bool)
	signalChan := make(chan os.Signal, 10)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	readers := make([]*FileTailReader, 0, 50)
	for _, fc := range conf.Files {
		readers = append(readers, NewReader(fc.Path, fc.AppType))
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, f := range conf.Files {
					for _, reader := range readers {
						reader.ApplyNewContent(func(v string) {
							onEntry(v, f.AppType)
						})
					}
				}
			case quit := <-quitChan:
				if quit {
					ticker.Stop()
					onStop()
				}
			case <-signalChan:
				log.Print("Caught signal, exiting...")
				ticker.Stop()
				onStop()
			}
		}
	}()
}
