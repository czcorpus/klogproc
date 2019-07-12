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
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

var (
	datetimePattern = regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2}\\s[012]\\d:[0-5]\\d:[0-5]\\d)[\\.,]\\d+")
)

type Conf struct {
	SrcPath                string `json:"srcPath"`
	PartiallyMatchingFiles bool   `json:"partiallyMatchingFiles"`
	WorklogPath            string `json:"worklogPath"`
}

// importTimeFromLine import a datetime information from the beginning
// of kontext applog. Because KonText does not log a timezone information
// it must be passed here to produce proper datetime.
//
// In case of an error, -1 is returned along with the error
func importTimeFromLine(lineStr string, timezoneStr string) (int64, error) {
	srch := datetimePattern.FindStringSubmatch(lineStr)
	var err error
	if len(srch) > 0 {
		if t, err := time.Parse("2006-01-02 15:04:05-07:00", srch[1]+timezoneStr); err == nil {
			return t.Unix(), nil
		}
	}
	return -1, err
}

// getFileMtime returns file's UNIX mtime (in secods).
// In case of an error, -1 is returned
func getFileMtime(filePath string) int64 {
	f, err := os.Open(filePath)
	if err != nil {
		return -1
	}
	finfo, err := f.Stat()
	if err == nil {
		return finfo.ModTime().Unix()
	}
	return -1
}

// LogFileMatches tests whether the log file specified by filePath matches
// in terms of its first record (whether it is older than the 'minTimestamp').
// If strictMatch is false then in case of non matching file, also its mtime
// is tested.
//
// The function expects that the first line on any log file contains proper
// log record which should be OK (KonText also writes multi-line error dumps
// to the log but it always starts with a proper datetime information).
func LogFileMatches(filePath string, minTimestamp int64, strictMatch bool, timezoneStr string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	rd := bufio.NewScanner(f)
	rd.Scan()
	line := rd.Text()
	startTime, err := importTimeFromLine(line, timezoneStr)
	if err != nil {
		return false, err
	}

	if startTime < minTimestamp && !strictMatch {
		startTime = getFileMtime(filePath)
	}

	return startTime >= minTimestamp, nil
}

// getFilesInDir lists all the matching log files
func getFilesInDir(dirPath string, minTimestamp int64, strictMatch bool, timezoneStr string) []string {
	tmp, err := ioutil.ReadDir(dirPath)
	var ans []string
	if err == nil {
		ans = make([]string, len(tmp))
		i := 0
		for _, item := range tmp {
			logPath := path.Join(dirPath, item.Name())
			matches, merr := LogFileMatches(logPath, minTimestamp, strictMatch, timezoneStr)
			if merr != nil {
				log.Println("ERROR: Failed to check log file ", logPath)

			} else if matches {
				ans[i] = logPath
				i++
			}
		}
		return ans[:i]
	}
	return []string{}
}

// LogItemProcessor is an object handling individual
type LogItemProcessor interface {
	ProcItem(logRec conversion.InputRecord) conversion.OutputRecord
	GetAppType() string
}

// LogFileProcFunc is a function for batch/tail processing of file-based logs
type LogFileProcFunc = func(conf *Conf, localTimezone string, minTimestamp int64)

// CreateLogFileProcFunc joins a defined log transformer and output channels to and
// returns a customized function for file/directory processing.
func CreateLogFileProcFunc(processor LogItemProcessor, destChans ...chan conversion.OutputRecord) LogFileProcFunc {
	return func(conf *Conf, localTimezone string, minTimestamp int64) {
		files := getFilesInDir(conf.SrcPath, minTimestamp, !conf.PartiallyMatchingFiles, localTimezone)
		log.Printf("Found %d file(s) to process in %s", len(files), conf.SrcPath)
		for _, file := range files {
			p := newParser(file, localTimezone, processor.GetAppType())
			p.Parse(minTimestamp, processor, destChans...)
		}
	}
}
