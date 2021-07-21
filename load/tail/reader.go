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
	"io"
	"log"
	"os"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/fsop"
)

// FileTailReader reads newly added lines to a file.
// Important assumptions:
// 1) file changes only by appending new lines
// 2) during normal operation, the inode of the file remains the same
// 3) change of inode means we start reading a new file from the beginning
type FileTailReader struct {
	processor    FileTailProcessor
	internalSeek int64
	file         *os.File
	filePath     string
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
func (ftw *FileTailReader) ApplyNewContent(processor FileTailProcessor, prevPosition conversion.LogRange) error {
	currInode, _, err := fsop.GetFileProps(processor.FilePath())
	if err != nil {
		return err
	}
	newPosition := conversion.LogRange{SeekEnd: -1, Inode: currInode}
	if currInode != prevPosition.Inode {
		ftw.internalSeek = 0
		ftw.file.Close()
		ftw.file, err = os.Open(ftw.processor.FilePath())
		if err != nil {
			return err
		}

	} else if !prevPosition.Written {
		ftw.internalSeek = prevPosition.SeekStart
		log.Printf("WARNING: FileTailReader(%s) updated internalSeek position to %d due to unsaved last record", ftw.filePath, prevPosition.SeekStart)
		ftw.file.Seek(ftw.internalSeek, io.SeekStart)

	} else if ftw.internalSeek != prevPosition.SeekEnd {
		// some external action has changed processed position (typically in case of a write error)
		if ftw.internalSeek == -1 {
			ftw.file.Close()
			ftw.file, err = os.Open(ftw.processor.FilePath())
			if err != nil {
				return err
			}
		}
		ftw.internalSeek = prevPosition.SeekEnd
		ftw.file.Seek(ftw.internalSeek, io.SeekStart)
		log.Printf("WARNING: FileTailReader[%s] updated internalSeek position to %d due to updated position status", ftw.filePath, ftw.internalSeek)
	}

	sc := bufio.NewReader(ftw.file)
	for {
		newPosition.SeekStart = ftw.internalSeek
		rawLine, err := sc.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		newPosition.SeekEnd = newPosition.SeekStart + int64(len(rawLine))
		ftw.internalSeek = newPosition.SeekEnd
		processor.OnEntry(string(rawLine[:len(rawLine)-1]), newPosition)
	}
	return nil
}

// NewReader creates a new file reader instance
func NewReader(processor FileTailProcessor, lastLogPosition conversion.LogRange) (*FileTailReader, error) {
	r := &FileTailReader{
		processor:    processor,
		internalSeek: -1, // this triggers initial read
		file:         nil,
		filePath:     processor.FilePath(),
	}
	if lastLogPosition.Inode > 0 {
		var err error
		r.file, err = os.Open(processor.FilePath())
		if err != nil {
			return nil, err
		}
		_, err = r.file.Seek(lastLogPosition.SeekEnd, os.SEEK_SET)
		if err != nil {
			return nil, err
		}

	}
	return r, nil
}
