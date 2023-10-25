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

package celery

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"klogproc/save/elastic"
	"klogproc/save/influx"
	"klogproc/servicelog"
	"klogproc/servicelog/celery"

	"github.com/rs/zerolog/log"
)

const (
	defaultTickerIntervalSecs = 30
)

// AppConf specifies a single
type AppConf struct {
	Workdir string `json:"workdir"`
	Name    string `json:"name"`
}

// Conf specifies CeleryStatus module configuration
type Conf struct {
	CeleryBinaryPath string    `json:"celeryBinaryPath"`
	IntervalSecs     int       `json:"intervalSecs"`
	Apps             []AppConf `json:"apps"`
}

// ---------

func sliceContains(items []int, v int) bool {
	for _, w := range items {
		if v == w {
			return true
		}
	}
	return false
}

// ---------------

type Processor struct {
	lastPIDs    []int
	reader      *StatusReader
	transformer *celery.Transformer
}

func (p *Processor) calcNumPIDChanges(rec *celery.InputRecord) int {

	var numChanges int
	for _, np := range rec.Pool.Processes {
		if !sliceContains(p.lastPIDs, np) {
			numChanges++
		}
	}
	return numChanges
}

func (p *Processor) Process() (*celery.OutputRecord, error) {
	rec, err := p.reader.ReadStatus()
	if err != nil {
		return nil, err
	}
	out, err := p.transformer.Transform(rec)
	out.NumWorkerRestarts = p.calcNumPIDChanges(rec)
	p.lastPIDs = rec.Pool.Processes
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Run controls regular checking of one or more Celery apps
func Run(conf *Conf, finishEvent chan<- bool, influxConf *influx.ConnectionConf, esConf *elastic.ConnectionConf,
	confirmChan chan interface{}) {
	tickerInterval := time.Duration(conf.IntervalSecs)
	if tickerInterval == 0 {
		log.Warn().Msgf("intervalSecs for Celery status mode not set, using default %ds", defaultTickerIntervalSecs)
		tickerInterval = time.Duration(defaultTickerIntervalSecs)

	} else {
		log.Info().Msgf("configured to check Celery status every %d second(s)", tickerInterval)
	}
	ticker := time.NewTicker(tickerInterval * time.Second)
	quitChan := make(chan bool)
	syscallChan := make(chan os.Signal)
	signal.Notify(syscallChan, os.Interrupt)
	signal.Notify(syscallChan, syscall.SIGTERM)
	processors := make([]*Processor, len(conf.Apps))
	for i, rc := range conf.Apps {
		processors[i] = &Processor{
			reader:      NewStatusReader(conf.CeleryBinaryPath, &rc),
			transformer: &celery.Transformer{},
		}
	}
	for {
		select {
		case <-ticker.C:
			var readSync sync.WaitGroup
			readSync.Add(len(processors))
			saveChannelInflux := make(chan *servicelog.BoundOutputRecord, influxConf.PushChunkSize)
			saveChannelES := make(chan *servicelog.BoundOutputRecord, esConf.PushChunkSize)
			var writeSync sync.WaitGroup
			writeSync.Add(2)
			go influx.RunWriteConsumer(influxConf, saveChannelInflux) // nil => no need to synchronize with other stuff
			go elastic.RunWriteConsumer("celery", esConf, saveChannelES)

			for _, proc := range processors {
				go func(proc *Processor) {
					out, err := proc.Process()
					if err != nil {
						log.Error().Err(err).Msgf("failed to process Celery status item")
					}
					saveChannelInflux <- &servicelog.BoundOutputRecord{Rec: out}
					saveChannelES <- &servicelog.BoundOutputRecord{Rec: out}
					readSync.Done()
				}(proc)
			}
			readSync.Wait()
			close(saveChannelInflux) // now the chunk save is triggered
			close(saveChannelES)
			writeSync.Wait()
		case quit := <-quitChan:
			if quit {
				ticker.Stop()
				finishEvent <- true
				return
			}
		case <-syscallChan:
			log.Info().Msg("Caught signal, exiting...")
			ticker.Stop()
			finishEvent <- true
			return
		}
	}
}
