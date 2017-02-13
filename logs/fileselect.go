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
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

func GetFilesInDir(dirPath string) []string {
	tmp, err := ioutil.ReadDir(dirPath)
	var ans []string
	if err == nil {
		ans = make([]string, len(tmp))
		for i, item := range tmp {
			ans[i] = path.Join(dirPath, item.Name())
		}
		return ans
	}
	return []string{}
}

func WorklogExists(s string) bool {
	_, err := os.Stat(s)
	return !os.IsNotExist(err)
}

type Worklog struct {
	reader *bufio.Scanner
	writer *bufio.Writer
}

func (w *Worklog) FindLastRecord() int {
	lastLine := ""
	if w.reader != nil {
		for w.reader.Scan() {
			lastLine = w.reader.Text()
		}
	}
	if lastLine != "" {
		ans, err := strconv.ParseInt(lastLine, 10, 32)
		if err == nil {
			return int(ans)
		}
	}
	return -1
}

func LoadWorklog(path string) (*Worklog, error) {
	var ans *Worklog
	var err error

	if WorklogExists(path) {
		f, err := os.Open(path)
		if err == nil {
			ans = &Worklog{reader: bufio.NewScanner(f)}

		} else {
			ans = &Worklog{}
		}

	} else {
		ans = &Worklog{}
		err = nil
	}
	return ans, err
}
