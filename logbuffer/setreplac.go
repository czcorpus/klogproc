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
	data []T
	cap  int
}

// Add adds a new value to the sample.
// It returns the sample size after the value was added
func (sample *SampleWithReplac[T]) Add(item T) int {
	if len(sample.data) < sample.cap {
		sample.data = append(sample.data, item)

	} else {
		idx := rand.Intn(sample.cap)
		sample.data[idx] = item
	}
	return len(sample.data)
}

func (sample SampleWithReplac[T]) Len() int {
	return len(sample.data)
}

func (sample SampleWithReplac[T]) GetAll() []T {
	return sample.data
}

func NewSampleWithReplac[T any](initialCap int) *SampleWithReplac[T] {
	return &SampleWithReplac[T]{
		data: make([]T, 0, initialCap),
		cap:  initialCap,
	}
}
