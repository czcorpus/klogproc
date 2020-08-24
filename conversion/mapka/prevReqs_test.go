// Copyright 2020 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2020 Institute of the Czech National Corpus,
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

package mapka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPrevReqPoolIncrement tests whether the position of the last
// element starts from zero again once it hits the pool size
func TestPrevReqPoolIncrement(t *testing.T) {
	p := PrevReqPool{}
	for i := 0; i < poolSize-1; i++ {
		p.AddItem(&OutputRecord{
			IPAddress: "127.0.0.1",
			UserAgent: "xx",
			Action:    "text",
		})
	}
	assert.Equal(t, 0, p.getFirstIdx())

	for i := 0; i < 11; i++ {
		p.AddItem(&OutputRecord{
			IPAddress: "127.0.0.1",
			UserAgent: "xx",
			Action:    "text",
		})
	}
	assert.Equal(t, 11, p.getFirstIdx())
}

func TestPoolSearchNoPropMatch(t *testing.T) {
	p := NewPrevReqPool(60)
	tested := &OutputRecord{
		ID:        "r1",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0...",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-10)),
	}
	p.AddItem(&OutputRecord{
		ID:        "r2",
		IPAddress: "192.168.1.10",
		UserAgent: "Mozilla/5.0...",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-10)),
	})
	p.AddItem(&OutputRecord{
		ID:        "r3",
		IPAddress: "127.0.0.1",
		UserAgent: "Different browser",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-10)),
	})
	p.AddItem(&OutputRecord{
		ID:        "r1",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0...",
		Action:    "overlay",
		time:      time.Now().Add(time.Second * time.Duration(-10)),
	})

	assert.False(t, p.ContainsSimilar(tested))
}

func TestPoolSearchTooOld(t *testing.T) {
	p := NewPrevReqPool(5)
	p.AddItem(&OutputRecord{
		ID:        "r1",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0...",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-10)),
	})

	tested := &OutputRecord{
		ID:        "r1",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0...",
		Action:    "text",
		time:      time.Now(),
	}

	assert.False(t, p.ContainsSimilar(tested))
}

func TestPoolSearchFindsProperMatch(t *testing.T) {
	p := NewPrevReqPool(9)
	p.AddItem(&OutputRecord{
		ID:        "r2",
		IPAddress: "192.168.1.10",
		UserAgent: "Mozilla/5.0...",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-15)),
	})
	p.AddItem(&OutputRecord{
		ID:        "r3",
		IPAddress: "127.0.0.1",
		UserAgent: "Different browser",
		Action:    "text",
		time:      time.Now().Add(time.Second * time.Duration(-8)),
	})
	p.AddItem(&OutputRecord{
		ID:        "r1",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0...",
		Action:    "overlay",
		time:      time.Now().Add(time.Second * time.Duration(-3)),
	})

	tested := &OutputRecord{
		ID:        "r500",
		IPAddress: "127.0.0.1",
		UserAgent: "Different browser",
		Action:    "text",
		time:      time.Now(),
	}

	assert.True(t, p.ContainsSimilar(tested))
}
