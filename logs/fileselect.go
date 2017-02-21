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

package logs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

// importTimeFromLine import a datetime information from the beginning
// of kontext applog. Because KonText does not log a timezone information
// it must be passed here to produce proper datetime.
//
// In case of an error, -1 is returned
func importTimeFromLine(lineStr string, timezoneStr string) int64 {
	rg := regexp.MustCompile("^(\\d{4}-\\d{2}-\\d{2}\\s[012]\\d:[0-5]\\d:[0-5]\\d)[\\.,]\\d+")
	srch := rg.FindStringSubmatch(lineStr)
	if len(srch) > 0 {
		if t, err := time.Parse("2006-01-02 15:04:05-07:00", srch[1]+timezoneStr); err == nil {
			return t.Unix()
		}
	}
	return -1
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
	startTime := importTimeFromLine(line, timezoneStr)

	if startTime < minTimestamp && !strictMatch {
		startTime = getFileMtime(filePath)
	}

	return startTime >= minTimestamp, nil
}

// GetFilesInDir lists all the matching log files
func GetFilesInDir(dirPath string, minTimestamp int64, strictMatch bool, timezoneStr string) []string {
	tmp, err := ioutil.ReadDir(dirPath)
	var ans []string
	if err == nil {
		ans = make([]string, len(tmp))
		i := 0
		for _, item := range tmp {
			logPath := path.Join(dirPath, item.Name())
			matches, merr := LogFileMatches(logPath, minTimestamp, strictMatch, timezoneStr)
			if merr != nil {
				log.Println("Failed to check log file ", logPath)

			} else if matches {
				ans[i] = logPath
				i++
			}
		}
		return ans[:i]
	}
	return []string{}
}

// ------------------------ worklog ---------

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
				lastLine = reader.Text()
			}
		}
		if lastLine != "" {
			ans, err := strconv.ParseInt(lastLine, 10, 64)
			if err == nil {
				return ans
			}
		}

	} else {
		log.Println("No worklog file present - all the found log files will be processed")
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

// NewWorklog creates an instance of Worklog with
// defined path. No file access attempts are made.
func NewWorklog(path string) *Worklog {
	return &Worklog{filePath: path}
}
