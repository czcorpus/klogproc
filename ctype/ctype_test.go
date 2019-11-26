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

package ctype

import (
	"testing"

	"github.com/czcorpus/klogproc/conversion/kontext"
	"github.com/stretchr/testify/assert"
)

func TestAgentIsBot(t *testing.T) {
	rec := &kontext.InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: kontext.Request{
			HTTPUserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		},
	}
	analyzer := &ClientTypeAnalyzer{
		data: &BotsAndMonitors{
			bots: []BotInfo{
				BotInfo{
					title:   "",
					match:   []string{"Googlebot/", "Mozilla/5.0"},
					example: "",
				},
			},
		},
	}
	assert.True(t, analyzer.AgentIsBot(rec))
}

func TestAgentIsBotMustMatchAll(t *testing.T) {
	rec := &kontext.InputRecord{
		Date: "2019-06-25T14:04:50.23-01:00",
		Request: kontext.Request{
			HTTPUserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		},
	}
	analyzer := &ClientTypeAnalyzer{
		data: &BotsAndMonitors{
			bots: []BotInfo{
				BotInfo{
					title:   "",
					match:   []string{"Googlebot/", "Mozilla/6.0"},
					example: "",
				},
			},
		},
	}
	assert.False(t, analyzer.AgentIsBot(rec))
}
