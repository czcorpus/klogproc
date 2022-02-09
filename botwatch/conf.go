// Copyright 2022 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2022 Institute of the Czech National Corpus,
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

package botwatch

import (
	"klogproc/botwatch/redisdb"
)

// BotDetection defines important parameters of bot detection module which determine
// how dense and regular request activity is considered as scripted/bot-like.
type BotDetectionConf struct {

	// BotDefsPath is either a local filesystem path or http resource path
	// where a list of bots to ignore etc. is defined
	BotDefsPath string `json:"botDefsPath"`

	// WatchedTimeWindowSecs specifies a time interval during which IP activies are evaluated.
	// In other words - each new record is considered along with older records at most as old
	// as specified by this property
	WatchedTimeWindowSecs int `json:"watchedTimeWindowSecs"`

	// NumRequestsThreshold specifies how many requests must be present during
	// WatchedTimeWindowSecs to treat the series as "bot-like"
	NumRequestsThreshold int `json:"numRequestsThreshold"`

	// RSDThreshold is a relative standard deviation (aka Coefficient of variation)
	// threshold of subsequent request intervals considered as bot-like
	RSDThreshold float64 `json:"rsdThreshold"`

	Redis redisdb.ConnectionConf `json:"redis"`
}
