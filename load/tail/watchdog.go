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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"klogproc/conversion"
	"klogproc/load"
	"klogproc/save"

	"github.com/czcorpus/cnc-gokit/fs"
	"github.com/rs/zerolog/log"
)

const (
	defaultTickerIntervalSecs = 60
)

// FileConf represents a configuration for a single
// log file to be watched
type FileConf struct {
	Path    string `json:"path"`
	AppType string `json:"appType"`
	// Version represents a major and minor version signature as used in semantic versioning
	// (e.g. 0.15, 1.2)
	Version string           `json:"version"`
	TZShift int              `json:"tzShift"`
	Buffer  *load.BufferConf `json:"buffer"`
}

func (fc *FileConf) Validate() error {
	if pathExists := fs.PathExists(fc.Path); !pathExists {
		return fmt.Errorf("failed to validate FileConf for %s - path does not exist	", fc.Path)
	}
	return nil
}

func (fc *FileConf) GetPath() string {
	return fc.Path
}

func (fc *FileConf) GetAppType() string {
	return fc.AppType
}

// Conf wraps all the configuration for the 'tail' function
type Conf struct {
	IntervalSecs          int        `json:"intervalSecs"`
	MaxLinesPerCheck      int        `json:"maxLinesPerCheck"`
	WorklogPath           string     `json:"worklogPath"`
	Files                 []FileConf `json:"files"`
	NumErrorsAlarm        int        `json:"numErrorsAlarm"`
	ErrCountTimeRangeSecs int        `json:"errCountTimeRangeSecs"`
}

// FullFiles provides a slice of `FileConf` with items where
// only Buffer.ID is filled upgraded to full config. This
// solves situations where user wants to share
// buffer between file processors and the buffer is configured
// only for one of the processors (which is reasonable as
// otherwise, there would be quite lot of rendundant conf. data)
func (conf *Conf) FullFiles() ([]FileConf, error) {
	buffConfs := make(map[string]*load.BufferConf)
	for _, v := range conf.Files {
		if v.Buffer != nil && v.Buffer.HasConfiguredBufferProcessing() && v.Buffer.IsShared() {
			buffConfs[v.Buffer.ID] = v.Buffer
		}
	}
	ans := make([]FileConf, len(conf.Files))
	for i, v := range conf.Files {
		ans[i] = v
		if v.Buffer != nil && v.Buffer.IsShared() && !v.Buffer.HasConfiguredBufferProcessing() {
			conf, ok := buffConfs[v.Buffer.ID]
			if !ok {
				return []FileConf{}, fmt.Errorf(
					"invalid shared buffer ID %s - full conf. not found", v.Buffer.ID)
			}
			ans[i].Buffer = conf
		}
	}
	return ans, nil
}

func (conf *Conf) RequiresMailConfiguration() bool {
	return conf.NumErrorsAlarm > 0 && conf.ErrCountTimeRangeSecs > 0
}

func (conf *Conf) Validate() error {
	if conf.IntervalSecs < 10 {
		return errors.New("logTail.intervalSecs must be at least 10")
	}
	if conf.MaxLinesPerCheck < conf.IntervalSecs*100 {
		return errors.New("logTail.maxLinesPerCheck must be at least logTail.intervalSecs * 100")
	}
	for _, fc := range conf.Files {
		if err := fc.Validate(); err != nil {
			return fmt.Errorf("logTail.files validation error: %w", err)
		}
	}
	return nil
}

type LineProcConfirmChan chan interface{}

// LogDataWriter represents a per-check scoped instance
// for writing converted logs to respective databases.
// I.e. in case checks overlap due to a too long processing
// in the previous check, both runs can independently write
// their data.
type LogDataWriter struct {
	Elastic chan *conversion.BoundOutputRecord
	Influx  chan *conversion.BoundOutputRecord
	Ignored chan save.IgnoredItemMsg
}

// FileTailProcessor specifies an object which is able to utilize all
// the "events" watchdog provides when processing a file tail for
// a concrete appType
type FileTailProcessor interface {
	AppType() string
	FilePath() string
	MaxLinesPerCheck() int
	CheckIntervalSecs() int

	// OnCheckStart marks start of logged file check
	// it returns a writer for storing converted adata
	// and also a channel where confirmations of writes
	// are sent.
	OnCheckStart() (LineProcConfirmChan, *LogDataWriter)

	// OnEntry is called on each processed line
	OnEntry(writer *LogDataWriter, item string, logPosition conversion.LogRange)

	// OnCheckStop marks the end of the single file check
	OnCheckStop(writer *LogDataWriter)
	OnQuit()
}

func initReaders(processors []FileTailProcessor, worklog *Worklog) ([]*FileTailReader, error) {
	readers := make([]*FileTailReader, len(processors))
	for i, processor := range processors {
		wlItem := worklog.GetData(processor.FilePath())
		log.Info().Msgf("Found log file %s", processor.FilePath())
		if wlItem.Inode > -1 {
			log.Info().Msgf("Found worklog for %s: %v", processor.FilePath(), wlItem)

		} else {
			log.Warn().Msgf("no worklog for %s - creating a new one...", processor.FilePath())
			inode, err := worklog.ResetFile(processor.FilePath())
			if err != nil {
				return readers, err
			}
			log.Info().Msgf("... added a worklog record for %s, inode: %d", processor.FilePath(), inode)
		}
		rdr, err := NewReader(
			processor,
			worklog.GetData(processor.FilePath()),
		)
		if err != nil {
			return readers, err
		}
		readers[i] = rdr
	}
	return readers, nil
}

// Run starts the process of (multiple) log watching
func Run(conf *Conf, processors []FileTailProcessor, finishEvent chan<- bool) {
	tickerInterval := time.Duration(conf.IntervalSecs)
	if tickerInterval == 0 {
		log.Warn().Msgf("intervalSecs for tail mode not set, using default %ds", defaultTickerIntervalSecs)
		tickerInterval = time.Duration(defaultTickerIntervalSecs)

	} else {
		log.Info().Msgf("configured to check for file changes every %d second(s)", tickerInterval)
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
		log.Error().Err(err).Msg("")
		quitChan <- true

	} else {
		readers, err = initReaders(processors, worklog)
		if err != nil {
			log.Error().Err(err).Msg("")
			quitChan <- true
		}
	}

	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup
			wg.Add(len(readers))
			for _, reader := range readers {
				go func(rdr *FileTailReader) {
					actionChan, writer := rdr.Processor().OnCheckStart()
					go func() {
						for action := range actionChan {
							switch action := action.(type) {
							case save.ConfirmMsg:
								if action.Error != nil {
									log.Error().Err(action.Error).Msg("Failed to write data to one of target databases")
								}
								worklog.UpdateFileInfo(action.FilePath, action.Position)
							case save.IgnoredItemMsg:
								worklog.UpdateFileInfo(action.FilePath, action.Position)
							}
						}
						wg.Done()
					}()
					prevPos := worklog.GetData(rdr.processor.FilePath())
					rdr.ApplyNewContent(rdr.Processor(), writer, prevPos)
					rdr.Processor().OnCheckStop(writer)
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
			log.Warn().Msg("Caught signal, exiting...")
			ticker.Stop()
			for _, reader := range readers {
				reader.Processor().OnQuit()
			}
			worklog.Close()
			finishEvent <- true
		}
	}
}
