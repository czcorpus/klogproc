package scripting

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

var (
	ErrScriptingNotSupported = errors.New("scripting not supported")
	ErrFailedTypeAssertion   = errors.New("failed type assertion")
)

type InvalidAttrError struct {
	Attr string
}

func (err InvalidAttrError) Error() string {
	return fmt.Sprintf("error accessing attribute '%s'", err.Attr)
}

func LuaTableToSliceOfStrings(val *lua.LTable) ([]string, error) {
	tableSize := val.Len()
	ans := make([]string, tableSize)
	for i := 1; i <= tableSize; i++ { // note: Lua tables are 1-based
		v := val.RawGetInt(i)
		if tv, ok := v.(lua.LString); ok {
			ans[i-1] = string(tv)

		} else {
			return ans, ErrFailedTypeAssertion
		}
	}
	return ans, nil
}

func LuaTableToMap(val *lua.LTable) map[string]any {
	ans := make(map[string]any)
	val.ForEach(func(key, val lua.LValue) {
		if tKey, ok := key.(lua.LString); ok {
			switch tVal := val.(type) {
			case lua.LString:
				ans[string(tKey)] = string(tVal)
			case lua.LNumber:
				ans[string(tKey)] = float64(tVal)
			case lua.LBool:
				ans[string(tKey)] = tVal == lua.LTrue
			case *lua.LTable:
				ans[string(tKey)] = LuaTableToMap(tVal)
			}
		}
	})
	return ans
}

func LuaTableToSliceOfInts(val *lua.LTable) ([]int, error) {
	tableSize := val.Len()
	ans := make([]int, tableSize)
	for i := 1; i <= tableSize; i++ { // note: Lua tables are 1-based
		v := val.RawGetInt(i)
		if tv, ok := v.(lua.LNumber); ok {
			ans[i-1] = int(tv)

		} else {
			return ans, ErrFailedTypeAssertion
		}
	}
	return ans, nil
}

func LuaTableToSliceOfFloats(val *lua.LTable) ([]float64, error) {
	tableSize := val.Len()
	ans := make([]float64, tableSize)
	for i := 1; i <= tableSize; i++ { // note: Lua tables are 1-based
		v := val.RawGetInt(i)
		if tv, ok := v.(lua.LNumber); ok {
			ans[i-1] = float64(tv)

		} else {
			return ans, ErrFailedTypeAssertion
		}
	}
	return ans, nil
}

func LuaTableToSliceOfBools(val *lua.LTable) ([]bool, error) {
	tableSize := val.Len()
	ans := make([]bool, tableSize)
	for i := 1; i <= tableSize; i++ { // note: Lua tables are 1-based
		v := val.RawGetInt(i)
		ans[i-1] = v == lua.LTrue // here we cannot use type assertion (see Gopher-lua docs)
	}
	return ans, nil
}
