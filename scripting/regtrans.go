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
	"fmt"
	"klogproc/servicelog"

	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

const (
	transformerMTName = "default_transformer_mt"
	transformerName   = "default_transformer"
)

func checkTransformer(L *lua.LState, pos int) servicelog.LogItemTransformer {
	ud := L.CheckUserData(pos)
	if v, ok := ud.Value.(servicelog.LogItemTransformer); ok {
		return v
	}
	L.ArgError(1, "servicelog.LogItemTransformer expected")
	return nil
}

func registerStaticTransformer[T servicelog.LogItemTransformer](env *lua.LState, transformer T) error {

	transFn := func(L *lua.LState) int {
		lrec := L.CheckUserData(2)
		fmt.Println("LREC: ", lrec)
		tLrec, ok := lrec.Value.(servicelog.InputRecord)
		if !ok {
			L.ArgError(2, "expected InputRecord")
		}
		tzshift := L.CheckInt(3)
		ans, err := transformer.Transform(tLrec, tzshift)
		if err != nil {
			L.RaiseError("failed to transform record: %s", err)
		}
		ud := env.NewUserData()
		ud.Value = ans
		env.SetMetatable(ud, env.GetTypeMetatable(outputRecName))
		env.Push(ud)
		return 1
	}

	var transformerMethods = map[string]lua.LGFunction{
		"transform": transFn,
	}
	mt := env.NewTypeMetatable(transformerMTName)
	env.SetGlobal(transformerMTName, mt)
	env.SetGlobal("set_out", env.NewFunction(func(e *lua.LState) int {
		orec := checkOutputRecord(env, 1)
		key := env.CheckString(2)
		val := env.CheckAny(3)
		if err := transformer.SetOutputProperty(orec, key, val); err != nil {
			// TODO
			log.Error().Err(err).Msg("failed to set output property")
		}
		return 0
	}))
	env.SetField(mt, "__index", env.SetFuncs(env.NewTable(), transformerMethods))
	tt := env.NewUserData()
	env.SetMetatable(tt, mt)
	env.SetGlobal(transformerName, tt)
	return nil
}
