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
	"klogproc/servicelog"

	lua "github.com/yuin/gopher-lua"
)

func checkOutputRecord(L *lua.LState, pos int) servicelog.OutputRecord {
	ud := L.CheckUserData(pos)
	if v, ok := ud.Value.(servicelog.OutputRecord); ok {
		return v
	}
	L.ArgError(1, "servicelog.OutputRecord expected")
	return nil
}

func registerOutputRecord(env *lua.LState, outRecFact func() servicelog.OutputRecord) {
	env.SetGlobal("new_out_rec", env.NewFunction(func(env *lua.LState) int {
		ud := env.NewUserData()
		ud.Value = outRecFact()
		env.Push(ud)
		return 1
	}))
}
