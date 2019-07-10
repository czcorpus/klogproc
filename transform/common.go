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

package transform

import (
	"fmt"
	"net"
	"time"

	"github.com/czcorpus/klogproc/transform/kontext"
)

type InputRecord interface {
	GetTime() time.Time
	GetClientIP() net.IP
	AgentIsLoggable() bool
}

type LogTransformer interface {
	ProcItem(appType string, logRec InputRecord) *kontext.OutputRecord
}

func Run(input InputRecord) (*kontext.OutputRecord, error) {
	switch tInput := input.(type) {
	case *kontext.InputRecord:
		return kontext.New(tInput, "kontext"), nil
	default:
		return nil, fmt.Errorf("Unsupported input record type %T", input)
	}
}
