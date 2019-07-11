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

package batch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFilesInDir(t *testing.T) {
	rootDir, err := os.Getwd()
	if err != nil {
		t.Fail()
	}

	// this should cause the function to return only two latest log files
	limit := int64(1485887057)
	// TODO we can test realiably only strict mode
	files := getFilesInDir(filepath.Join(rootDir, "..", "..", "testdata", "logs"), limit, true, "+01:00")
	if len(files) != 2 {
		t.Errorf("Invalid number of files detected - expected 2, found %d ", len(files))
	}
}
