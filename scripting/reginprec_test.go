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

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

func TestFieldAccess(t *testing.T) {
	lt := new(ltrans)
	exe, err := CreateCustomTransformer(
		"print(\"megatest\")\n"+
			"function transform (rec)\n"+
			"  return rec.ClientIP\n"+
			"end\n",
		lt,
		func(env *lua.LState) {},
	)
	if assert.NoError(t, err) {
		env := exe.GetLState()
		registerInputRecord(exe.GetLState())

		inp := &dummyInputRec{
			ID:       "foobar",
			ClientIP: net.IPv4(192, 168, 1, 10),
		}

		env.Push(env.GetGlobal("transform"))
		env.Push(importInputRecord(env, inp))
		err = env.PCall(1, 1, nil)
		assert.NoError(t, err)
		ret := env.Get(-1) // returned value
		env.Pop(1)
		tRet, ok := ret.(lua.LString)
		assert.True(t, ok)
		assert.Equal(t, "192.168.1.10", string(tRet))
	}

}
