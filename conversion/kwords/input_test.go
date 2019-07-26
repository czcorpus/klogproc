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

package kwords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTime(t *testing.T) {
	rec := InputRecord{
		Datetime:  "2019-06-25T14:04:50.23-01:00",
		IPAddress: "192.168.1.65",
	}
	m := rec.GetTime()
	assert.Equal(t, 2019, m.Year())
	assert.Equal(t, 6, int(m.Month()))
	assert.Equal(t, 25, m.Day())
	assert.Equal(t, 14, m.Hour())
	assert.Equal(t, 4, m.Minute())
	assert.Equal(t, 50, m.Second())
	_, d := m.Zone()
	assert.Equal(t, -3600, d)
}

func TestGetIPAddress(t *testing.T) {
	rec := InputRecord{
		IPAddress: "192.168.1.65",
	}
	a := rec.GetClientIP()
	assert.Equal(t, "192.168.1.65", a.String())
}

func TestAgentIsLoggable(t *testing.T) {
	rec := InputRecord{}
	assert.True(t, rec.AgentIsLoggable())
}
