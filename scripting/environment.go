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

	lua "github.com/yuin/gopher-lua"
)

func prepareScript(env *lua.LState, srcPath string) error {
	if err := env.DoFile(srcPath); err != nil {
		return fmt.Errorf("failed to process customization script %s: %w", srcPath, err)
	}
	return nil
}

func CreateCustomTransformer(sourceCode string, transformer servicelog.LogItemTransformer, beforeRun func(env *lua.LState)) (*Transformer, error) {
	scriptingEngine := lua.NewState()
	beforeRun(scriptingEngine)
	if err := scriptingEngine.DoString(sourceCode); err != nil {
		return nil, fmt.Errorf("failed to process customization source code: %w", err)
	}
	return &Transformer{env: scriptingEngine, staticTransformer: transformer}, nil
}

func CreateEnvironment(
	logConf servicelog.LogProcConf,
	defaultTransformer servicelog.LogItemTransformer,
	outRecFactory func() servicelog.OutputRecord,
) (*lua.LState, error) {
	L := lua.NewState()
	registerInputRecord(L)
	registerOutputRecord(L, outRecFactory)
	registerProcConf(L, logConf)
	registerStaticTransformer(L, defaultTransformer)
	SetupRequireFn(L)
	if err := prepareScript(L, logConf.GetScriptPath()); err != nil {
		return nil, fmt.Errorf("failed to process script %s: %w", logConf.GetScriptPath(), err)
	}
	return L, nil
}

// ------------------------------------

type Transformer struct {
	env               *lua.LState
	staticTransformer servicelog.LogItemTransformer
	anonymousUsers    []int
}

func (e *Transformer) GetLState() *lua.LState {
	return e.env
}

func (t *Transformer) HistoryLookupItems() int {
	return t.staticTransformer.HistoryLookupItems()
}

func (t *Transformer) AppType() string {
	return t.staticTransformer.AppType()
}

func (t *Transformer) Preprocess(rec servicelog.InputRecord, prevRecs servicelog.ServiceLogBuffer) []servicelog.InputRecord {
	return t.staticTransformer.Preprocess(rec, prevRecs)
}

func (t *Transformer) Transform(logRec servicelog.InputRecord, tzShiftMin int) (servicelog.OutputRecord, error) {
	if t.env != nil {
		aus := &lua.LTable{}
		for i, v := range t.anonymousUsers {
			aus.RawSetInt(i+1, lua.LNumber(v))
		}
		fnObj := t.env.GetGlobal("transform")
		if fnObj == lua.LNil {
			return nil, fmt.Errorf(
				"failed to transform record of type %s using a Lua script: missing `transform` function",
				t.AppType())
		}
		err := t.env.CallByParam(
			lua.P{
				Fn:      fnObj,
				NRet:    1,
				Protect: true,
			},
			importInputRecord(t.env, logRec), lua.LNumber(tzShiftMin),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to transform record of type %s using a Lua script: %w", t.AppType(), err)
		}
		ret := t.env.Get(-1)
		tRet, ok := ret.(*lua.LUserData)
		if !ok {
			return nil, fmt.Errorf(
				"failed to transform record of type %s using a Lua script: assertion error",
				t.AppType())
		}
		t.env.Pop(1)
		unwrapped, ok := tRet.Value.(servicelog.OutputRecord)
		if !ok {
			return nil, fmt.Errorf(
				"failed to transform record of type %s using a Lua script: invalid type of wrapped value",
				t.AppType())
		}
		return unwrapped, nil

	}
	return t.staticTransformer.Transform(logRec, tzShiftMin)
}

func (t *Transformer) SetOutputProperty(rec servicelog.OutputRecord, name string, value lua.LValue) error {
	return t.staticTransformer.SetOutputProperty(rec, name, value)
}

func NewTransformer(env *lua.LState, staticTransformer servicelog.LogItemTransformer) *Transformer {
	return &Transformer{env: env, staticTransformer: staticTransformer}
}
