package scripting

import (
	"errors"
	"fmt"
	"reflect"

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

func StructToLua(L *lua.LState, val reflect.Value) (*lua.LTable, error) {
	table := L.NewTable()

	// go from pointer to the original value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, ErrFailedTypeAssertion
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		if !fieldType.IsExported() {
			continue
		}

		// Convert the field value to a Lua value
		var lValue lua.LValue
		switch field.Kind() {
		case reflect.String:
			lValue = lua.LString(field.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			lValue = lua.LNumber(field.Int())
		case reflect.Float32, reflect.Float64:
			lValue = lua.LNumber(field.Float())
		case reflect.Bool:
			lValue = lua.LBool(field.Bool())
		case reflect.Slice, reflect.Array:
			sliceTable := L.NewTable()
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct {
					nestedStruct, err := StructToLua(L, elem)
					if err != nil {
						return nil, fmt.Errorf("failed to convert struct to Lua: %w", err)
					}
					L.SetTable(sliceTable, lua.LNumber(j+1), nestedStruct)
				} else {
					// Handle primitive types in slices
					switch elem.Kind() {
					case reflect.String:
						L.SetTable(sliceTable, lua.LNumber(j+1), lua.LString(elem.String()))
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						L.SetTable(sliceTable, lua.LNumber(j+1), lua.LNumber(elem.Int()))
					case reflect.Float32, reflect.Float64:
						L.SetTable(sliceTable, lua.LNumber(j+1), lua.LNumber(elem.Float()))
					case reflect.Bool:
						L.SetTable(sliceTable, lua.LNumber(j+1), lua.LBool(elem.Bool()))
					}
				}
			}
			lValue = sliceTable
		case reflect.Struct:
			var err error
			lValue, err = StructToLua(L, field)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
		L.SetTable(table, lua.LString(fieldType.Name), lValue)
	}

	return table, nil
}
