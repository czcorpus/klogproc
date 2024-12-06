// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package scripting

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

func TestFieldAccess(t *testing.T) {
	lt := new(ltrans)
	exe, err := CreateCustomTransformer(
		"function transform (rec)\n"+
			"  return {rec.ClientIP, rec.Time, rec.Args.Name}\n"+
			"end\n",
		lt,
		func(env *lua.LState) {},
	)
	if assert.NoError(t, err) {
		L := exe.GetLState()
		registerInputRecord(exe.GetLState())
		tz, err := time.LoadLocation("Europe/Prague")
		assert.NoError(t, err)
		inp := &dummyInputRec{
			ID:       "foobar",
			Time:     time.Date(2024, time.December, 1, 10, 28, 13, 0, tz),
			ClientIP: net.IPv4(192, 168, 1, 10),
			Args: struct {
				Name     string
				Position int
			}{
				Name:     "arg1",
				Position: 1000,
			},
		}

		L.Push(L.GetGlobal("transform"))
		L.Push(importInputRecord(L, inp))
		err = L.PCall(1, 1, nil)
		assert.NoError(t, err)
		ret := L.Get(-1) // returned value
		L.Pop(1)
		tRet, ok := ret.(*lua.LTable)
		assert.True(t, ok)
		assert.Equal(t, 3, tRet.Len())

		rawRes := tRet.RawGetInt(1)
		v1, ok := rawRes.(lua.LString)
		assert.True(t, ok)
		assert.Equal(t, "192.168.1.10", string(v1))
		rawRes = tRet.RawGetInt(2)
		v2, ok := rawRes.(lua.LString)
		assert.True(t, ok)
		assert.Equal(t, "2024-12-01T10:28:13+01:00", string(v2))
	}

}
