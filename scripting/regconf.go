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

const (
	appConfName = "conf"
)

/*
here we push tile processing configuration so Lua scripts can read it
*/

func registerProcConf(env *lua.LState, confProvider servicelog.LogProcConf) {

	val := new(lua.LTable)
	val.RawSetString("appType", lua.LString(confProvider.GetAppType()))
	exclIP := new(lua.LTable)
	val.RawSetString("excludeIPList", exclIP)
	for i, ip := range confProvider.GetExcludeIPList() {
		exclIP.RawSetInt(i+1, lua.LString(ip))
	}
	val.RawSetString("version", lua.LString(confProvider.GetVersion()))
	env.SetGlobal(appConfName, val)
}
