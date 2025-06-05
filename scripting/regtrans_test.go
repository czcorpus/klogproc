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
	"klogproc/logbuffer"
	"klogproc/servicelog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// Mock buffer implementation
type mockBuffer struct{}

func (mb *mockBuffer) AddRecord(rec servicelog.InputRecord) {}
func (mb *mockBuffer) ConfirmRecordCheck(rec logbuffer.Storable) {}
func (mb *mockBuffer) GetLastCheck(clusteringID string) time.Time { return time.Now() }
func (mb *mockBuffer) RemoveAnalyzedRecords(clusteringID string, dt time.Time) {}
func (mb *mockBuffer) NumOfRecords(clusteringID string) int { return 0 }
func (mb *mockBuffer) ClearOldRecords(maxAge time.Time) int { return 0 }
func (mb *mockBuffer) TotalNumOfRecordsSince(dt time.Time) int { return 0 }
func (mb *mockBuffer) GetRecords(clusteringID string, from time.Time) []servicelog.InputRecord { return nil }
func (mb *mockBuffer) ForEach(clusteringID string, fn func(item servicelog.InputRecord)) {}
func (mb *mockBuffer) TotalForEach(fn func(item servicelog.InputRecord)) {}
func (mb *mockBuffer) SetStateData(stateData logbuffer.SerializableState) {}
func (mb *mockBuffer) GetStateData(dtNow time.Time) logbuffer.SerializableState { return nil }
func (mb *mockBuffer) EmptyStateData() logbuffer.SerializableState { return nil }

func setupLuaState() (*lua.LState, *ltrans) {
	L := lua.NewState()
	transformer := &ltrans{}
	err := registerStaticTransformer(L, transformer)
	if err != nil {
		panic(err)
	}
	return L, transformer
}

func TestTransformDefault(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock input record
	inputRec := &dummyInputRec{
		ID:        "test-123",
		Time:      time.Date(2024, 6, 5, 12, 30, 0, 0, time.UTC),
		ClientIP:  net.ParseIP("192.168.1.1"),
		UserAgent: "Mozilla/5.0",
	}

	// Create user data for input record
	ud := L.NewUserData()
	ud.Value = inputRec

	// Test transform_default function
	err := L.DoString(`
		function test_transform()
			local out = transform_default(input_rec)
			return out
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("input_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_transform"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result
	result := L.Get(-1)
	assert.Equal(t, lua.LTUserData, result.Type())
	
	outRec, ok := result.(*lua.LUserData).Value.(servicelog.OutputRecord)
	assert.True(t, ok)
	assert.Equal(t, "test-123", outRec.GetID())
}

func TestPreprocessDefault(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock input record
	inputRec := &dummyInputRec{
		ID:        "test-456",
		Time:      time.Date(2024, 6, 5, 14, 15, 0, 0, time.UTC),
		ClientIP:  net.ParseIP("10.0.0.1"),
		UserAgent: "TestAgent/1.0",
	}

	// Create mock buffer
	buffer := &mockBuffer{}

	// Create user data
	udInput := L.NewUserData()
	udInput.Value = inputRec
	udBuffer := L.NewUserData()
	udBuffer.Value = buffer

	// Test preprocess_default function
	err := L.DoString(`
		function test_preprocess()
			local result = preprocess_default(input_rec, buffer)
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("input_rec", udInput)
	L.SetGlobal("buffer", udBuffer)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_preprocess"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result is a table (slice converted to Lua table)
	result := L.Get(-1)
	assert.Equal(t, lua.LTTable, result.Type())
}

func TestSetOutProp(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record
	outRec := &dummyOutRec{
		ID:          "test-789",
		Time:        time.Now().Format(time.RFC3339),
		IsAI:        false,
		CustomField: "original",
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test set_out_prop function - this should fail since dummyOutRec doesn't support LSetProperty
	err := L.DoString(`
		function test_set_prop()
			set_out_prop(output_rec, "CustomField", "modified")
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_set_prop"),
		NRet:    0,
		Protect: true,
	})
	// Should fail because dummyOutRec returns ErrScriptingNotSupported
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "set_out_prop failed")
}

func TestGetOutProp(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record
	outRec := &dummyOutRec{
		ID:          "test-get-prop",
		Time:        time.Now().Format(time.RFC3339),
		IsAI:        true,
		CustomField: "test-value",
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test get_out_prop function
	err := L.DoString(`
		function test_get_prop()
			local val = get_out_prop(output_rec, "IsAI")
			return val
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_get_prop"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result
	result := L.Get(-1)
	assert.Equal(t, lua.LTBool, result.Type())
	assert.Equal(t, lua.LTrue, result)
}

func TestDatetimeAddMinutes(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	originalTime := time.Date(2024, 6, 5, 12, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-datetime",
		time: originalTime,
		Time: originalTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test datetime_add_minutes function - add 30 minutes
	err := L.DoString(`
		function test_datetime_add()
			datetime_add_minutes(output_rec, 30)
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_datetime_add"),
		NRet:    0,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check that time was modified
	expectedTime := originalTime.Add(30 * time.Minute)
	assert.Equal(t, expectedTime, outRec.GetTime())
	assert.Equal(t, expectedTime.Format(time.RFC3339), outRec.Time)
}

func TestDatetimeAddMinutesNegative(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	originalTime := time.Date(2024, 6, 5, 12, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-datetime-neg",
		time: originalTime,
		Time: originalTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test datetime_add_minutes function - subtract 45 minutes
	err := L.DoString(`
		function test_datetime_subtract()
			datetime_add_minutes(output_rec, -45)
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_datetime_subtract"),
		NRet:    0,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check that time was modified
	expectedTime := originalTime.Add(-45 * time.Minute)
	assert.Equal(t, expectedTime, outRec.GetTime())
	assert.Equal(t, expectedTime.Format(time.RFC3339), outRec.Time)
}

func TestIsBeforeDateTime(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	recordTime := time.Date(2024, 6, 5, 10, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-before",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_before_datetime function - record time is before comparison time
	err := L.DoString(`
		function test_is_before_true()
			local result = is_before_datetime(output_rec, "2024-06-05T12:00:00")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_before_true"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result is true (record time 10:00 is before 12:00)
	result := L.Get(-1)
	assert.Equal(t, lua.LTBool, result.Type())
	assert.Equal(t, lua.LTrue, result)
}

func TestIsBeforeDateTimeFalse(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	recordTime := time.Date(2024, 6, 5, 14, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-before-false",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_before_datetime function - record time is after comparison time
	err := L.DoString(`
		function test_is_before_false()
			local result = is_before_datetime(output_rec, "2024-06-05T12:00:00")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_before_false"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result is false (record time 14:00 is not before 12:00)
	result := L.Get(-1)
	assert.Equal(t, lua.LTBool, result.Type())
	assert.Equal(t, lua.LFalse, result)
}

func TestIsBeforeDateTimeInvalidFormat(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record
	recordTime := time.Date(2024, 6, 5, 10, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-before-invalid",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_before_datetime function with invalid datetime format
	err := L.DoString(`
		function test_is_before_invalid()
			local result = is_before_datetime(output_rec, "invalid-datetime")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_before_invalid"),
		NRet:    1,
		Protect: true,
	})
	// Should fail due to invalid datetime format
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse datetime argument")
}

func TestIsAfterDateTime(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	recordTime := time.Date(2024, 6, 5, 14, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-after",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_after_datetime function - record time is after comparison time
	err := L.DoString(`
		function test_is_after_true()
			local result = is_after_datetime(output_rec, "2024-06-05T12:00:00")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_after_true"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result is true (record time 14:00 is after 12:00)
	result := L.Get(-1)
	assert.Equal(t, lua.LTBool, result.Type())
	assert.Equal(t, lua.LTrue, result)
}

func TestIsAfterDateTimeFalse(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record with specific time
	recordTime := time.Date(2024, 6, 5, 10, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-after-false",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_after_datetime function - record time is before comparison time
	err := L.DoString(`
		function test_is_after_false()
			local result = is_after_datetime(output_rec, "2024-06-05T12:00:00")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_after_false"),
		NRet:    1,
		Protect: true,
	})
	assert.NoError(t, err)

	// Check result is false (record time 10:00 is not after 12:00)
	result := L.Get(-1)
	assert.Equal(t, lua.LTBool, result.Type())
	assert.Equal(t, lua.LFalse, result)
}

func TestIsAfterDateTimeInvalidFormat(t *testing.T) {
	L, _ := setupLuaState()
	defer L.Close()

	// Create mock output record
	recordTime := time.Date(2024, 6, 5, 14, 0, 0, 0, time.UTC)
	outRec := &dummyOutRec{
		ID:   "test-after-invalid",
		time: recordTime,
		Time: recordTime.Format(time.RFC3339),
	}

	// Create user data
	ud := L.NewUserData()
	ud.Value = outRec

	// Test is_after_datetime function with invalid datetime format
	err := L.DoString(`
		function test_is_after_invalid()
			local result = is_after_datetime(output_rec, "invalid-datetime")
			return result
		end
	`)
	assert.NoError(t, err)

	L.SetGlobal("output_rec", ud)
	err = L.CallByParam(lua.P{
		Fn:      L.GetGlobal("test_is_after_invalid"),
		NRet:    1,
		Protect: true,
	})
	// Should fail due to invalid datetime format
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse datetime argument")
}