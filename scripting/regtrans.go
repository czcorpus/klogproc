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

	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

func registerStaticTransformer[T servicelog.LogItemTransformer](env *lua.LState, transformer T) error {

	env.SetGlobal("transform_default", env.NewFunction(func(L *lua.LState) int {
		lrec := L.CheckUserData(1)
		tLrec, ok := lrec.Value.(servicelog.InputRecord)
		if !ok {
			L.ArgError(1, "expected InputRecord")
		}
		tzshift := L.CheckInt(2)
		ans, err := transformer.Transform(tLrec, tzshift)
		if err != nil {
			L.RaiseError("failed to transform record: %s", err)
		}
		ud := env.NewUserData()
		ud.Value = ans
		env.Push(ud)
		return 1
	}))

	env.SetGlobal("set_out_prop", env.NewFunction(func(e *lua.LState) int {
		orec := checkOutputRecord(env, 1)
		key := env.CheckString(2)
		val := env.CheckAny(3)
		if err := transformer.SetOutputProperty(orec, key, val); err != nil {
			// TODO
			log.Error().Err(err).Msg("failed to set output property")
		}
		return 0
	}))

	return nil
}
