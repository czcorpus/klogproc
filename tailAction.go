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

package main

import (
	"log"
	"sync"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/load/batch"
	"github.com/czcorpus/klogproc/load/tail"
	"github.com/oschwald/geoip2-golang"
)

type notifyFailedChunks struct{}

func (n *notifyFailedChunks) RescueFailedChunks(chunk [][]byte) error {
	if len(chunk) > 0 {
		log.Print("ERROR: failed to insert a chunk of size ", len(chunk))
	}
	return nil
}

// -----

type tailProcessor struct {
	conf             *Conf
	lineParsers      map[string]batch.LineParser
	logTransformers  map[string]conversion.LogItemTransformer
	geoDB            *geoip2.Reader
	dataForES        chan conversion.OutputRecord
	dataForInflux    chan conversion.OutputRecord
	localTimezone    string
	anonymousUsers   []int
	elasticChunkSize int
	influxChunkSize  int
	outSync          sync.WaitGroup
}

func (tp *tailProcessor) OnCheckStart() {
	tp.dataForES = make(chan conversion.OutputRecord, tp.elasticChunkSize*2)
	tp.dataForInflux = make(chan conversion.OutputRecord, tp.influxChunkSize)
	tp.outSync = sync.WaitGroup{}
	tp.outSync.Add(2)
	go runElasticWrite(tp.conf, tp.dataForES, &tp.outSync, &notifyFailedChunks{})
	go runInfluxWrite(tp.conf, tp.dataForInflux, &tp.outSync)
}

func (tp *tailProcessor) OnEntry(item, appType string) {
	parsed, err := tp.lineParsers[appType].ParseLine(item, 0, tp.localTimezone)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return
	}
	outRec, err := tp.logTransformers[appType].Transform(parsed, appType, tp.anonymousUsers)
	if err != nil {
		log.Printf("ERROR: %s", err)
		return
	}
	applyLocation(parsed, tp.geoDB, outRec)
	tp.dataForES <- outRec
	tp.dataForInflux <- outRec
}

func (tp *tailProcessor) OnCheckStop() {
	close(tp.dataForES)
	close(tp.dataForInflux)
	tp.outSync.Wait()
}

func (tp *tailProcessor) OnQuit() {

}

func newTailProcessor(conf *Conf, geoDB *geoip2.Reader) *tailProcessor {
	var err error

	lineParsers := make(map[string]batch.LineParser)
	logTransformers := make(map[string]conversion.LogItemTransformer)
	for _, f := range conf.LogTail.Files {
		lineParsers[f.AppType], err = batch.NewLineParser(f.AppType)
		if err != nil {
			log.Fatal("ERROR: Failed to initialize parser: ", err)
		}
		logTransformers[f.AppType], err = GetLogTransformer(f.AppType)
		if err != nil {
			log.Fatal("ERROR: Failed to initialize transformer: ", err)
		}
	}

	return &tailProcessor{
		conf:             conf,
		lineParsers:      lineParsers,
		logTransformers:  logTransformers,
		geoDB:            geoDB,
		localTimezone:    conf.LocalTimezone,
		anonymousUsers:   conf.AnonymousUsers,
		elasticChunkSize: conf.ElasticSearch.PushChunkSize,
		influxChunkSize:  conf.InfluxDB.PushChunkSize,
	}
}

// -----

func runTailAction(conf *Conf, geoDB *geoip2.Reader, finishEvt chan bool) {
	proc := newTailProcessor(conf, geoDB)
	tail.Run(&conf.LogTail, proc, finishEvt)
}
