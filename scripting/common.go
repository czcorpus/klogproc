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

func LuaTableToSliceOfUserData[T any](val *lua.LTable) ([]T, error) {
	tableSize := val.Len()
	ans := make([]T, tableSize)
	for i := 1; i <= tableSize; i++ { // note: Lua tables are 1-based
		v := val.RawGetInt(i)
		if ud, ok := v.(*lua.LUserData); ok {
			if typedItem, ok := ud.Value.(T); ok {
				ans[i-1] = typedItem
				continue
			}
		}
		return ans, ErrFailedTypeAssertion
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

func ValueToLua(L *lua.LState, val reflect.Value) (lua.LValue, error) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// Convert the field value to a Lua value
	var lValue lua.LValue
	switch val.Kind() {
	case reflect.String:
		lValue = lua.LString(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		lValue = lua.LNumber(val.Int())
	case reflect.Float32, reflect.Float64:
		lValue = lua.LNumber(val.Float())
	case reflect.Bool:
		lValue = lua.LBool(val.Bool())
	case reflect.Slice, reflect.Array:
		sliceTable := L.NewTable()
		for j := 0; j < val.Len(); j++ {
			elem := val.Index(j)
			if elem.Kind() == reflect.Struct {
				nestedStruct, err := StructToLua(L, elem)
				if err != nil {
					return lua.LNil, fmt.Errorf("failed to convert struct to Lua: %w", err)
				}
				L.SetTable(sliceTable, lua.LNumber(j+1), nestedStruct)
			} else {
				// Handle primitive types in slices, everything else will be LUserData
				switch elem.Kind() {
				case reflect.String:
					L.SetTable(sliceTable, lua.LNumber(j+1), lua.LString(elem.String()))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					L.SetTable(sliceTable, lua.LNumber(j+1), lua.LNumber(elem.Int()))
				case reflect.Float32, reflect.Float64:
					L.SetTable(sliceTable, lua.LNumber(j+1), lua.LNumber(elem.Float()))
				case reflect.Bool:
					L.SetTable(sliceTable, lua.LNumber(j+1), lua.LBool(elem.Bool()))
				default:
					ud := L.NewUserData()
					ud.Value = elem.Interface()
					L.SetTable(sliceTable, lua.LNumber(j+1), ud)
				}
			}
		}
		lValue = sliceTable
	case reflect.Struct:
		var err error
		lValue, err = StructToLua(L, val)
		if err != nil {
			return lua.LNil, err
		}
	default:
		lValue = lua.LNil
	}
	return lValue, nil
}

func StructPropToLua(L *lua.LState, val reflect.Value, propName string) (lua.LValue, error) {
	// go from pointer to the original value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return lua.LNil, ErrFailedTypeAssertion
	}
	fld := val.FieldByName(propName)
	if fld.IsValid() {
		return ValueToLua(L, fld)
	}
	return lua.LNil, fmt.Errorf("invalid property %s of object %s", propName, val.Type().Name())
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

		lValue, err := ValueToLua(L, field)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to convert struct of type %s to Lua: %w", val.Type().Name(), err)
		}
		L.SetTable(table, lua.LString(fieldType.Name), lValue)
	}

	return table, nil
}
