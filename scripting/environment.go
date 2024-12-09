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
	"errors"
	"fmt"
	"klogproc/servicelog"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

func prepareScript(env *lua.LState, srcPath string) error {
	if err := env.DoFile(srcPath); err != nil {
		return fmt.Errorf("failed to process customization script %s: %w", srcPath, err)
	}
	return nil
}

func CreateCustomTransformer(sourceCode string, transformer servicelog.LogItemTransformer, beforeRun func(env *lua.LState)) (*Transformer, error) {
	L := lua.NewState()
	beforeRun(L)
	if err := L.DoString(sourceCode); err != nil {
		return nil, fmt.Errorf("failed to process customization source code: %w", err)
	}
	return &Transformer{L: L, staticTransformer: transformer}, nil
}

// testIRecProp tests whether a property of an InputRecord exists
func testIRecProp(L *lua.LState) int {
	ud := L.CheckUserData(1)
	name := L.CheckString(2)

	switch inputRec := ud.Value.(type) {
	case servicelog.InputRecord, servicelog.OutputRecord:
		val := reflect.ValueOf(inputRec).Elem()
		field := val.FieldByName(name)
		if field.IsValid() { // AFAIK we cannot use type conversion here
			L.Push(lua.LTrue)

		} else {
			L.Push(lua.LFalse)
		}
	default:
		L.ArgError(1, "expected input or output record")
	}

	return 1
}

func mkLogFn(logevtFact func() *zerolog.Event) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		msg := L.CheckString(1)
		logevt := logevtFact()
		if L.GetTop() == 2 {
			props := L.CheckTable(2)
			for k, v := range LuaTableToMap(props) {
				logevt = logevt.Any(k, v)
			}
		}
		logevt.Msg(msg)
		return 0
	}
}

func setupLogging(L *lua.LState) {
	logTbl := L.NewTable()
	logTblMt := L.NewTable()
	logTbl.Metatable = logTblMt
	L.SetField(
		logTblMt,
		"__index",
		L.SetFuncs(
			L.NewTable(),
			map[string]lua.LGFunction{
				"debug": mkLogFn(func() *zerolog.Event { return log.Debug() }),
				"info":  mkLogFn(func() *zerolog.Event { return log.Info() }),
				"warn":  mkLogFn(func() *zerolog.Event { return log.Warn() }),
				"error": func(L *lua.LState) int {
					msg := L.CheckString(1)
					logevt := log.Error()
					if L.GetTop() == 2 {
						props := L.CheckTable(2)
						for k, v := range LuaTableToMap(props) {
							logevt = logevt.Any(k, v)
						}
					}
					logevt.Err(errors.New(msg)).Send()
					return 0
				},
			},
		),
	)
	L.SetGlobal("logger", logTbl)
}

func CreateEnvironment(
	logConf servicelog.LogProcConf,
	anonymousUsers []int,
	defaultTransformer servicelog.LogItemTransformer,
	outRecFactory func() servicelog.OutputRecord,
) (*lua.LState, error) {
	L := lua.NewState()
	registerInputRecord(L)
	registerOutputRecord(L, outRecFactory)
	registerStaticTransformer(L, defaultTransformer)
	setupLogging(L)
	SetupRequireFn(L)

	L.SetGlobal("rec_prop_exists", L.NewFunction(testIRecProp))
	L.SetGlobal("app_type", lua.LString(logConf.GetAppType()))
	L.SetGlobal("app_version", lua.LString(logConf.GetVersion()))
	anUsers, err := ValueToLua(L, reflect.ValueOf(anonymousUsers))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Lua environment: %w", err)
	}
	L.SetGlobal("anonymous_users", anUsers)

	if err := prepareScript(L, logConf.GetScriptPath()); err != nil {
		return nil, fmt.Errorf("failed to process script %s: %w", logConf.GetScriptPath(), err)
	}
	return L, nil
}
