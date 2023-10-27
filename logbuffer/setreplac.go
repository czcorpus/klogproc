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
	"math/rand"
)

type SampleWithReplac[T any] struct {
	Data []T `json:"data"`
	Cap  int `json:"cap"`
}

// Add adds a new value to the sample.
// It returns the sample size after the value was added
func (sample *SampleWithReplac[T]) Add(item T) int {
	if len(sample.Data) < sample.Cap {
		sample.Data = append(sample.Data, item)

	} else {
		idx := rand.Intn(sample.Cap)
		sample.Data[idx] = item
	}
	return len(sample.Data)
}

func (sample *SampleWithReplac[T]) Resize(newSize int) {
	if newSize > sample.Cap {
		sample.Cap = newSize

	} else if newSize < sample.Cap {
		sample.Cap = newSize
		sample.Data = sample.Data[:newSize]
	}
}

func (sample *SampleWithReplac[T]) Len() int {
	return len(sample.Data)
}

func (sample *SampleWithReplac[T]) GetAll() []T {
	return sample.Data
}

func NewSampleWithReplac[T any](initialCap int) *SampleWithReplac[T] {
	return &SampleWithReplac[T]{
		Data: make([]T, 0, initialCap),
		Cap:  initialCap,
	}
}
