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
	"os"

	"klogproc/fsop"
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
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
func (ftw *FileTailReader) ApplyNewContent(
	processor FileTailProcessor,
	dataWriter *LogDataWriter,
	prevPosition servicelog.LogRange,
) error {
	currInode, _, err := fsop.GetFileProps(processor.FilePath())
	if err != nil {
		return err
	}
	newPosition := servicelog.LogRange{SeekEnd: -1, Inode: currInode}
	if currInode != prevPosition.Inode {
		ftw.internalSeek = 0
		ftw.file.Close()
		ftw.file, err = os.Open(ftw.processor.FilePath())
		if err != nil {
			return err
		}

	} else if !prevPosition.Written {
		ftw.internalSeek = prevPosition.SeekStart
		log.Warn().Msgf("FileTailReader(%s) updated internalSeek position to %d due to unsaved last record", ftw.filePath, prevPosition.SeekStart)

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
		log.Warn().Msgf("FileTailReader[%s] updated internalSeek position to %d due to updated position status", ftw.filePath, ftw.internalSeek)
	}
	// always make sure the current position is OK (it can be off e.g. thanks
	// to using the buffered reader)
	ftw.file.Seek(ftw.internalSeek, io.SeekStart)

	sc := bufio.NewReader(ftw.file)
	var i int
	for i = 0; i < ftw.processor.MaxLinesPerCheck(); i++ {
		newPosition.SeekStart = ftw.internalSeek
		rawLine, err := sc.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		newPosition.SeekEnd = newPosition.SeekStart + int64(len(rawLine))
		ftw.internalSeek = newPosition.SeekEnd
		processor.OnEntry(dataWriter, string(rawLine[:len(rawLine)-1]), newPosition)
	}
	if i == ftw.processor.MaxLinesPerCheck() {
		log.Warn().
			Int("maxLinesPerCheck", ftw.processor.MaxLinesPerCheck()).
			Str("logFile", ftw.filePath).
			Str("name", ftw.AppType()).
			Msg("tail processor hit the maxLinesPerCheck limit")

	} else {
		log.Debug().
			Int("processedLines", i).
			Str("logFile", ftw.filePath).
			Str("name", ftw.AppType()).
			Msg("processed a chunk of lines")

	}
	return nil
}

// NewReader creates a new file reader instance
func NewReader(processor FileTailProcessor, lastLogPosition servicelog.LogRange) (*FileTailReader, error) {
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
		_, err = r.file.Seek(lastLogPosition.SeekEnd, io.SeekStart)
		if err != nil {
			return nil, err
		}

	}
	return r, nil
}
