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
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

//go:embed lua/*.lua
var luaScripts embed.FS

func LoadEmbeddedScript(L *lua.LState, scriptPath string) error {
	content, err := fs.ReadFile(luaScripts, scriptPath)
	if err != nil {
		return fmt.Errorf("failed to load embedded script %s: %w", scriptPath, err)
	}
	if err := L.DoString(string(content)); err != nil {
		return fmt.Errorf("failed to execute embedded script %s: %w", scriptPath, err)
	}

	return nil
}

func SetupRequireFn(L *lua.LState) {
	packageTable := L.NewTable()
	L.SetGlobal("package", packageTable)
	loaded := L.NewTable()
	L.SetField(packageTable, "loaded", loaded)

	requireFn := L.NewFunction(func(L *lua.LState) int {
		modname := L.CheckString(1)
		if mod := L.GetField(loaded, modname); mod != lua.LNil {
			L.Push(mod)
			return 1
		}
		scriptPath := filepath.Join("lua", modname+".lua")
		content, err := luaScripts.ReadFile(scriptPath)
		if err != nil {
			L.RaiseError("embedded module `%s` not found: %v", modname, err)
			return 0
		}

		// Load the module
		fn, err := L.LoadString(string(content))
		if err != nil {
			L.RaiseError("error loading embedded module '%s': %v", modname, err)
			return 0
		}

		L.Push(fn)
		L.Call(0, 1)

		// Store the result in package.loaded
		moduleResult := L.Get(-1)
		L.SetField(loaded, modname, moduleResult)

		return 1
	})

	L.SetGlobal("require", requireFn)
}
