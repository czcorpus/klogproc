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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/czcorpus/klogproc/conversion/celery"
)

type nullWriter struct {
}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// ------------

// StatusReader fetches Celery status using os.exec and stores
// the JSON-like output to an InputRecord
type StatusReader struct {
	CeleryBinaryPath string
	AppName          string
	AppWorkdir       string
}

// ReadStatus reads current Celery status via cmd
func (s *StatusReader) ReadStatus() (*celery.InputRecord, error) {
	var stdOutput bytes.Buffer
	outWriter := bufio.NewWriter(&stdOutput)
	var errOutput bytes.Buffer
	errWriter := bufio.NewWriter(&errOutput)
	cmd := exec.Command(s.CeleryBinaryPath, "inspect", "stats", "--workdir", s.AppWorkdir, "-A", s.AppName)
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	err := cmd.Run()
	if err != nil {
		log.Print("WARNING: Celery inspect error output: ", string(errOutput.Bytes()))
		return nil, err
	}
	outWriter.Flush()
	ret := string(stdOutput.Bytes())
	idx := strings.Index(ret, "{")
	if idx >= 0 {
		var out celery.InputRecord
		err = json.Unmarshal([]byte(ret[idx:]), &out)
		if err != nil {
			return nil, err
		}
		return &out, nil
	}
	return nil, fmt.Errorf("Celery inspect response not recognized")
}

// NewStatusReader is a factory function for StatusReader
func NewStatusReader(celeryBinPath string, conf *AppConf) *StatusReader {
	return &StatusReader{
		CeleryBinaryPath: celeryBinPath,
		AppName:          conf.Name,
		AppWorkdir:       conf.Workdir,
	}
}
