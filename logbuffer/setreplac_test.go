// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package logbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	d := NewSampleWithReplac[int](5)
	assert.Equal(t, 0, d.Len())
	assert.Equal(t, 5, d.Cap)
}

func TestAddItems(t *testing.T) {
	d := NewSampleWithReplac[int](3)
	d.Add(1)
	d.Add(2)
	d.Add(3)
	assert.Equal(t, 3, d.Cap)
	assert.Equal(t, 3, d.Len())
	d.Add(100)
	assert.Equal(t, 3, d.Cap)
	assert.Equal(t, 3, d.Len())
}

func TestResizeInc(t *testing.T) {
	d := NewSampleWithReplac[int](3)
	d.Add(1)
	d.Add(2)
	d.Add(3)
	d.Resize(5)
	assert.Equal(t, 5, d.Cap)
	assert.Equal(t, 3, d.Len())
	d.Add(4)
	d.Add(5)
	d.Add(6)
	assert.Equal(t, 5, d.Cap)
	assert.Equal(t, 5, d.Len())
}

func TestResizeDec(t *testing.T) {
	d := NewSampleWithReplac[int](5)
	d.Add(1)
	d.Add(2)
	d.Add(3)
	d.Add(4)
	d.Add(5)
	d.Resize(3)
	assert.Equal(t, 3, d.Cap)
	assert.Equal(t, 3, d.Len())
	d.Add(6)
	assert.Equal(t, 3, d.Cap)
	assert.Equal(t, 3, d.Len())
}
