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

package shiny

import (
	"strconv"
	"time"

	"github.com/czcorpus/klogproc/conversion"
)

// Transformer converts a source log object into a destination one
type Transformer struct {
}

// Transform creates a new OutputRecord out of an existing InputRecord
func (t *Transformer) Transform(logRecord *InputRecord, recType string, anonymousUsers []int) (*OutputRecord, error) {
	userID := logRecord.User.ID
	if userID == 0 && len(anonymousUsers) > 0 {
		userID = anonymousUsers[0]
	}
	ans := &OutputRecord{
		Type:        recType,
		time:        logRecord.GetTime(),
		Datetime:    logRecord.GetTime().Format(time.RFC3339),
		IsQuery:     true,
		IPAddress:   logRecord.ClientIP,
		User:        logRecord.User.User,
		UserID:      strconv.Itoa(userID),
		IsAnonymous: conversion.UserBelongsToList(userID, anonymousUsers),
		Lang:        logRecord.Lang,
		UserAgent:   logRecord.UserAgent,
	}
	ans.ID = createID(ans)
	return ans, nil
}

func NewTransformer() *Transformer {
	return &Transformer{}
}
