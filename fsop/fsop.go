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

package fsop

import (
	"fmt"
	"os"
	"syscall"
)

// GetFileMtime returns file's UNIX mtime (in secods).
// In case of an error, -1 is returned
func GetFileMtime(filePath string) int64 {
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

// IsDir tests whether a provided path represents
// a directory. If not or in case of an IO error,
// false is returned.
func IsDir(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	finfo, err := f.Stat()
	if err != nil {
		return false
	}
	return finfo.Mode().IsDir()
}

// IsFile tests whether a provided path represents
// a file. If not or in case of an IO error,
// false is returned.
func IsFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	finfo, err := f.Stat()
	if err != nil {
		return false
	}
	return finfo.Mode().IsRegular()
}

func GetFileProps(filePath string) (inode int64, size int64, err error) {
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
