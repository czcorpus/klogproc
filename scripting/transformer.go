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
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

// ------------------------------------

type Transformer struct {
	L                 *lua.LState
	staticTransformer servicelog.LogItemTransformer
	anonymousUsers    []int
}

func (e *Transformer) GetLState() *lua.LState {
	return e.L
}

func (t *Transformer) HistoryLookupItems() int {
	return t.staticTransformer.HistoryLookupItems()
}

func (t *Transformer) AppType() string {
	return t.staticTransformer.AppType()
}

func (t *Transformer) Preprocess(
	rec servicelog.InputRecord,
	prevRecs servicelog.ServiceLogBuffer,
) ([]servicelog.InputRecord, error) {
	if t.L == nil {
		return t.staticTransformer.Preprocess(rec, prevRecs)
	}
	fnObj := t.L.GetGlobal("preprocess")
	if fnObj == lua.LNil {
		return nil, fmt.Errorf(
			"failed to preprocess record of type %s using a Lua script: missing `preprocess` function",
			t.AppType())
	}
	ud := t.L.NewUserData()
	ud.Value = prevRecs
	err := t.L.CallByParam(
		lua.P{
			Fn:      fnObj,
			NRet:    1,
			Protect: true,
		},
		importInputRecord(t.L, rec),
		ud,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute preprocess() function for application %s using a Lua script: %w", t.AppType(), err)
	}
	ret := t.L.Get(-1)
	tRet, ok := ret.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf(
			"failed to preprocess record of type %s using a Lua script - expected LUserData, got %s",
			t.AppType(), reflect.TypeOf(ret))
	}
	t.L.Pop(1)
	unwrapped, err := LuaTableToSliceOfUserData[servicelog.InputRecord](tRet)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to preprocess record of type %s using a Lua script: %s",
			t.AppType(), err)
	}
	return unwrapped, nil
}

func (t *Transformer) Transform(logRec servicelog.InputRecord) (servicelog.OutputRecord, error) {
	if t.L == nil {
		return t.staticTransformer.Transform(logRec)
	}
	aus := &lua.LTable{} // TODO what about anonymous users? should we pass it to transform?
	for i, v := range t.anonymousUsers {
		aus.RawSetInt(i+1, lua.LNumber(v))
	}
	fnObj := t.L.GetGlobal("transform")
	if fnObj == lua.LNil {
		return nil, fmt.Errorf(
			"failed to transform record of type %s using a Lua script: missing `transform` function",
			t.AppType())
	}
	err := t.L.CallByParam(
		lua.P{
			Fn:      fnObj,
			NRet:    1,
			Protect: true,
		},
		importInputRecord(t.L, logRec),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to transform record of type %s using a Lua script: %w", t.AppType(), err)
	}
	ret := t.L.Get(-1)
	tRet, ok := ret.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf(
			"failed to transform record of type %s using a Lua script: assertion error",
			t.AppType())
	}
	t.L.Pop(1)
	unwrapped, ok := tRet.Value.(servicelog.OutputRecord)
	if !ok {
		return nil, fmt.Errorf(
			"failed to transform record of type %s using a Lua script: invalid type of wrapped value",
			t.AppType())
	}
	return unwrapped, nil
}

func (t *Transformer) Close() {
	t.L.Close()
}

func NewTransformer(env *lua.LState, staticTransformer servicelog.LogItemTransformer) *Transformer {
	return &Transformer{L: env, staticTransformer: staticTransformer}
}
