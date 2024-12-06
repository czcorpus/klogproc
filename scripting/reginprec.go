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
	"net"
	"reflect"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	inputRecName = "input_rec_mt"
)

func importField(L *lua.LState, field reflect.Value) lua.LValue {
	switch field.Kind() {
	case reflect.String:
		return lua.LString(field.String())
	case reflect.Int, reflect.Int64:
		return lua.LNumber(float64(field.Int()))
	case reflect.Float64:
		return lua.LNumber(field.Float())
	case reflect.Bool:
		return lua.LBool(field.Bool())

	case reflect.Slice:
		if ipVal, ok := field.Interface().(net.IP); ok {
			return lua.LString(ipVal.String())
		}
		tbl := L.NewTable()
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			L.RawSetInt(tbl, i+1, importField(L, elem))
		}
		return tbl
	case reflect.Map:
		tbl := L.NewTable()
		iter := field.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := importField(L, iter.Value())
			L.RawSet(tbl, lua.LString(key.String()), value)
		}
		return tbl
	default:
		switch tVal := field.Interface().(type) {
		case time.Time:
			return lua.LString(tVal.Format(time.RFC3339))
		}
	}
	return lua.LNil
}

// getIRecProp
// TODO: this is not very effective for repeated calls on the same
// input report as each time, it creates a LFunction instance to be called
func getIRecProp(L *lua.LState, inputRec servicelog.InputRecord, name string) lua.LValue {
	val := reflect.ValueOf(inputRec).Elem()
	if field := val.FieldByName(name); field.IsValid() {
		return importField(L, field)
	}
	switch name {
	case "IsProcessable":
		return L.NewFunction(func(l *lua.LState) int {
			ans := inputRec.IsProcessable()
			if ans {
				L.Push(lua.LTrue)

			} else {
				L.Push(lua.LFalse)
			}
			return 1
		})
	case "ClusterSize":
		return L.NewFunction(func(l *lua.LState) int {
			ans := inputRec.ClusterSize()
			L.Push(lua.LNumber(ans))
			return 1
		})
	case "ClusteringClientID":
		return L.NewFunction(func(l *lua.LState) int {
			ans := inputRec.ClusteringClientID()
			L.Push(lua.LString(ans))
			return 1
		})
	case "IsSuspicious":
		return L.NewFunction(func(l *lua.LState) int {
			ans := inputRec.IsSuspicious()
			if ans {
				L.Push(lua.LTrue)
			} else {
				L.Push(lua.LFalse)
			}
			return 1
		})
	case "GetTime":
		return L.NewFunction(func(l *lua.LState) int {
			L.Push(lua.LString(inputRec.GetTime().Format(time.RFC3339)))
			return 1
		})
	}
	return lua.LNil
}

func get(L *lua.LState) int {
	irec := L.CheckUserData(1)
	tIrec, ok := irec.Value.(servicelog.InputRecord)
	if !ok {
		L.ArgError(1, "expecting InputRecord")
	}
	key := L.CheckString(2)
	ans := getIRecProp(L, tIrec, key)
	L.Push(ans)
	return 1
}

func importInputRecord(L *lua.LState, rec servicelog.InputRecord) lua.LValue {
	d := L.NewUserData()
	d.Value = rec
	L.SetMetatable(d, L.GetGlobal(inputRecName))
	return d
}

func registerInputRecord(L *lua.LState) {
	mt := L.NewTypeMetatable(inputRecName)
	L.SetGlobal(inputRecName, mt)
	L.SetField(mt, "__index", L.NewFunction(get))
}
