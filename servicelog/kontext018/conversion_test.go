// Copyright 2023 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
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

package kontext018

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

func TestSetPropsDatetime(t *testing.T) {
	tr := Transformer{}
	outRec := &OutputRecord{}
	tz, _ := time.LoadLocation("Europe/Prague")
	v := time.Date(2024, time.December, 2, 16, 51, 19, 0, tz)
	lv := lua.LString(v.Format(time.RFC3339))
	err := tr.SetOutputProperty(outRec, "Datetime", lv)
	assert.NoError(t, err)
	assert.Equal(t, "2024-12-02T16:51:19+01:00", outRec.Datetime)
	assert.Equal(t, v.Unix(), outRec.GetTime().Unix())
}
