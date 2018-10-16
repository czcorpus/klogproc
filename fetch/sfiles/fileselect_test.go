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

package sfiles

import (
	"testing"
)

func TestGetFilesInDir(t *testing.T) {
	limit := int64(1485887057)                                        // this should cause the function to return only two latest log files
	files := GetFilesInDir("../testdata/logs", limit, true, "+01:00") // TODO we can test realiably only strict mode
	if len(files) != 2 {
		t.Errorf("Invalid number of files detected - expected 2, found %d ", len(files))
	}
}
