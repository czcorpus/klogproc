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
	"net"
	"reflect"
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	inputRecName = "input_rec_mt"
)

func importField(L *lua.LState, field reflect.Value) lua.LValue {
	fmt.Println("KIND: ", field.Kind())
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

func getIRecProp(L *lua.LState, inputRec servicelog.InputRecord, name string) lua.LValue {
	val := reflect.ValueOf(inputRec).Elem()
	field := val.FieldByName(name)
	fmt.Println("FIELD ", field)
	if !field.IsValid() {
		return lua.LNil
	}
	return importField(L, field)
}

func get(env *lua.LState) int {
	irec := env.CheckUserData(1)
	tIrec, ok := irec.Value.(servicelog.InputRecord)
	if !ok {
		env.ArgError(1, "expecting InputRecord")
	}
	key := env.CheckString(2)
	fmt.Println("getting ", key, " FROM ", tIrec)
	ans := getIRecProp(env, tIrec, key)
	env.Push(ans)
	return 1
}

func registerInputRecord(env *lua.LState) {
	mt := env.NewTypeMetatable(inputRecName)
	env.SetGlobal(inputRecName, mt)
	env.SetField(mt, "__index", env.NewFunction(get))
}

func importInputRecord(L *lua.LState, rec servicelog.InputRecord) lua.LValue {
	d := L.NewUserData()
	d.Value = rec
	L.SetMetatable(d, L.GetGlobal(inputRecName))
	return d
}
