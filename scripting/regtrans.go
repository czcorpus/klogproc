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
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

func registerStaticTransformer[T servicelog.LogItemTransformer](
	L *lua.LState, transformer T,
) error {

	L.SetGlobal("transform_default", L.NewFunction(func(L *lua.LState) int {
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
		ud := L.NewUserData()
		ud.Value = ans
		L.Push(ud)
		return 1
	}))

	L.SetGlobal("preprocess_default", L.NewFunction(func(L *lua.LState) int {
		lrec := L.CheckUserData(1)
		tLrec, ok := lrec.Value.(servicelog.InputRecord)
		if !ok {
			L.ArgError(1, "expected InputRecord")
			return 1
		}
		logBuffer := L.CheckUserData(2)
		tLogBuffer, ok := logBuffer.Value.(servicelog.ServiceLogBuffer)
		if !ok {
			L.ArgError(1, "expected InputRecord")
			return 1
		}
		ans, err := transformer.Preprocess(tLrec, tLogBuffer)
		if err != nil {
			L.RaiseError("failed to preprocess(): %s", err)
		}
		lv, err := ValueToLua(L, reflect.ValueOf(ans))
		if err != nil {
			L.RaiseError("failed to run preprocess(): %s", err)
		}
		L.Push(lv)
		return 1
	}))

	L.SetGlobal("set_out_prop", L.NewFunction(func(e *lua.LState) int {
		orec := checkOutputRecord(L, 1)
		key := L.CheckString(2)
		val := L.CheckAny(3)
		if err := orec.LSetProperty(key, val); err != nil {
			L.RaiseError(
				"set_out_prop failed for type %s and key %s: %s",
				orec.GetType(), key, err,
			)
		}
		return 0
	}))

	L.SetGlobal("get_out_prop", L.NewFunction(func(e *lua.LState) int {
		orec := checkOutputRecord(L, 1)
		key := L.CheckString(2)
		val, err := StructPropToLua(L, reflect.ValueOf(orec), key)
		if err != nil {
			L.RaiseError("failed to get property: %s", err)
			return 0
		}
		L.Push(val)
		return 1
	}))

	return nil
}
