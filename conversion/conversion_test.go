// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertDatetimeString(t *testing.T) {
	m := ConvertDatetimeString("2019-06-25T14:04:50.23-01:00")
	assert.Equal(t, 2019, m.Year())
	assert.Equal(t, 6, int(m.Month()))
	assert.Equal(t, 25, m.Day())
	assert.Equal(t, 14, m.Hour())
	assert.Equal(t, 4, m.Minute())
	assert.Equal(t, 50, m.Second())
	_, d := m.Zone()
	assert.Equal(t, -3600, d)
}

func TestGetTimeInvalid(t *testing.T) {
	m := ConvertDatetimeString("total nonsense")
	assert.Equal(t, 1, m.Year())
}

func TestGetTimeNoTimezone(t *testing.T) {
	m := ConvertDatetimeString("2019-06-25T14:04:50.23")
	assert.Equal(t, 1, m.Year())
}

func TestImportBoolNumeric(t *testing.T) {
	b, err := ImportBool("1", "foo")
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = ImportBool("yes", "foo")
	assert.Nil(t, err)
	assert.True(t, b)

	b, err = ImportBool("0", "foo")
	assert.Nil(t, err)
	assert.False(t, b)

	b, err = ImportBool("no", "foo")
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestUserBelongsToList(t *testing.T) {
	assert.True(t, UserBelongsToList(37, []int{1, 2, 37, 38}))
	assert.False(t, UserBelongsToList(137, []int{1, 2, 37, 38}))
	assert.False(t, UserBelongsToList(0, []int{}))
}

func TestTimezoneToInt(t *testing.T) {
	mins, err := TimezoneToInt("-03:30")
	assert.Nil(t, err)
	assert.Equal(t, -210, mins)

	mins, err = TimezoneToInt("+02:00")
	assert.Nil(t, err)
	assert.Equal(t, 120, mins)

	mins, err = TimezoneToInt("08:30")
	assert.Error(t, err)

	mins, err = TimezoneToInt("a08:30")
	assert.Error(t, err)

	mins, err = TimezoneToInt("+a:b")
	assert.Error(t, err)

	mins, err = TimezoneToInt("+12-30")
	assert.Error(t, err)
}
