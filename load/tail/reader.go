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

// FileTailReader reads newly added lines to a file.
// Important assumptions:
// 1) file changes only by appending new lines
// 2) during normal operation, the inode of the file remains the same
// 3) change of inode means we start reading a new file from the beginning
type FileTailReader struct {
	processor   FileTailProcessor
	lastInode   int64
	lastSize    int64
	lastLine    int64
	file        *os.File
	filePath    string
	lastReadPos int64
}

// AppType returns app type identifier (kontext, syd, treq,...)
func (ftw *FileTailReader) AppType() string {
	return ftw.processor.AppType()
}

func (ftw *FileTailReader) FilePath() string {
	return ftw.filePath
}

// Processor returns attached file tail processor
func (ftw *FileTailReader) Processor() FileTailProcessor {
	return ftw.processor
}

// ApplyNewContent calls a provided function to newly added lines
func (ftw *FileTailReader) ApplyNewContent(onLine func(line string, lineNum int64), onDone func(inode int64, seek int64, lineNum int64)) error {
	currInode, currSize, err := getFileProps(ftw.processor.FilePath())
	if err != nil {
		return err
	}
	contentChanged := false

	if currInode != ftw.lastInode {
		contentChanged = true
		ftw.lastInode = currInode
		ftw.lastSize = currSize
		ftw.lastReadPos = 0
		ftw.lastLine = 0
		ftw.file.Close()
		ftw.file, err = os.Open(ftw.processor.FilePath())
		if err != nil {
			return err
		}

	} else if currSize != ftw.lastSize {
		contentChanged = true
	}

	if contentChanged {
		sc := bufio.NewScanner(ftw.file)
		for sc.Scan() {
			ftw.lastLine++
			onLine(sc.Text(), ftw.lastLine)
		}
		ftw.lastReadPos, err = ftw.file.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}
		onDone(ftw.lastInode, ftw.lastReadPos, ftw.lastLine)
	}
	return nil
}

// NewReader creates a new file reader instance
func NewReader(processor FileTailProcessor, lastInode, lastReadPos int64, lastLineNum int64) (*FileTailReader, error) {
	r := &FileTailReader{
		processor:   processor,
		lastInode:   lastInode,
		lastSize:    -1,
		lastLine:    lastLineNum,
		file:        nil,
		filePath:    processor.FilePath(),
		lastReadPos: lastReadPos,
	}
	if lastInode > 0 {
		var err error
		r.file, err = os.Open(processor.FilePath())
		if err != nil {
			return nil, err
		}
		_, err = r.file.Seek(lastReadPos, os.SEEK_SET)
		if err != nil {
			return nil, err
		}

	}
	return r, nil
}
