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
	"bufio"
	"fmt"
	"os"
	"syscall"
)

func getFileProps(filePath string) (inode int64, size int64, err error) {
	st, err := os.Stat(filePath)
	if err != nil {
		return -1, -1, err
	}
	stat, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return -1, -1, fmt.Errorf("Problem using syscall.Stat_t for file %s", filePath)
	}
	inode = int64(stat.Ino)
	size = st.Size()
	return
}

type FileTailReader struct {
	path        string
	lastInode   int64
	lastSize    int64
	file        *os.File
	lastReadPos int
}

func (ftw *FileTailReader) ApplyNewContent(callback func(line string)) error {
	currInode, currSize, err := getFileProps(ftw.path)
	if err != nil {
		return err
	}
	contentChanged := false

	if currInode != ftw.lastInode {
		contentChanged = true
		ftw.lastInode = currInode
		ftw.lastSize = currSize
		ftw.lastReadPos = 0
		ftw.file.Close()
		ftw.file, err = os.Open(ftw.path)
		if err != nil {
			return err
		}

	} else if currSize != ftw.lastSize {
		contentChanged = true
	}

	if contentChanged {
		sc := bufio.NewScanner(ftw.file)
		for sc.Scan() {
			callback(sc.Text())
		}
	}
	return nil
}

func NewReader(path string, appType string) *FileTailReader {
	return &FileTailReader{
		path:        path,
		lastInode:   -1,
		lastSize:    -1,
		file:        nil,
		lastReadPos: -1,
	}
}
