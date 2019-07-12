// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

package main

import (
	"fmt"

	"github.com/czcorpus/klogproc/conversion"
	"github.com/czcorpus/klogproc/conversion/kontext"
	"github.com/czcorpus/klogproc/conversion/syd"
)

// ------------------------------------

// konTextTransformer is a simple type-safe wrapper for kontext app type log transformer
type konTextTransformer struct {
	t *kontext.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (k *konTextTransformer) Transform(logRec conversion.InputRecord, recType string) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*kontext.InputRecord)
	if ok {
		return k.t.Transform(tRec, recType)
	}
	return nil, fmt.Errorf("Invalid type for conversion by KonText transformer %T", logRec)
}

// ------------------------------------

type sydTransformer struct {
	t *syd.Transformer
}

// Transform transforms KonText app log record types as general InputRecord
// In case of type mismatch, error is returned.
func (s *sydTransformer) Transform(logRec conversion.InputRecord, recType string) (conversion.OutputRecord, error) {
	tRec, ok := logRec.(*syd.InputRecord)
	if ok {
		return s.t.Transform(tRec, recType)
	}
	return nil, fmt.Errorf("Invalid type for conversion by SyD transformer %T", logRec)
}

// ------------------------------------

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(appType string) (conversion.LogItemTransformer, error) {

	switch appType {
	case conversion.AppTypeKontext:
		return &konTextTransformer{t: &kontext.Transformer{}}, nil
	case conversion.AppTypeSyd:
		return &sydTransformer{t: &syd.Transformer{}}, nil
	default:
		return nil, fmt.Errorf("Cannot find log transformer for app type %s", appType)
	}
}
