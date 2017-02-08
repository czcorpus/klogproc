// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tomachalek/klogproc/elastic"
	"github.com/tomachalek/klogproc/logs"
	"io/ioutil"
)

type Conf struct {
	WorklogPath   string                      `json:"worklogPath"`
	LogDir        string                      `json:"logDir"`
	ElasticServer string                      `json:"elasticServer"`
	ElasticIndex  string                      `json:"elasticIndex"`
	Updates       []elastic.APIFlagUpdateConf `json:"updates"`
}

func processLogs(conf *Conf) {
	worklog, err := logs.LoadWorklog(conf.WorklogPath)
	if err != nil {
		panic(err)
	}
	last := worklog.FindLastRecord()
	fmt.Println(worklog, last)
	files := logs.GetFilesInDir(conf.LogDir)
	fmt.Println("FILES: ", files)
	for _, file := range files {
		p := logs.NewParser(file)
		p.Parse(last)
	}
}

func updateIsAPIStatus(conf *Conf) {
	client := elastic.NewClient(conf.ElasticServer, conf.ElasticIndex)
	for _, updConf := range conf.Updates {
		ans, err := client.UpdateSetAPIFlag(updConf)
		fmt.Printf("CLIENT resp - ans: [%s], err: [%s]\n", ans, err)
	}
}

func loadConfig(path string) *Conf {
	rawData, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	var conf Conf
	json.Unmarshal(rawData, &conf)
	return &conf
}

func showHelp() {
	fmt.Println(`
Available operations: setapiflag, proclogs, help.
...TODO...`)
}

func main() {
	flag.Parse()

	if len(flag.Args()) == 1 && flag.Arg(0) == "help" {
		showHelp()

	} else if len(flag.Args()) == 2 {
		conf := loadConfig(flag.Arg(1))
		switch flag.Arg(0) {
		case "setapiflag":
			updateIsAPIStatus(conf)
		case "proclogs":
			processLogs(conf)
		}

	} else {
		panic("Invalid arguments. Expected format: klogproc OPERATION CONF")
	}

}
