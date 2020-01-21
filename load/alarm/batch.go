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

// fileselect functions are used to find proper KonText application log files
// based on logs processed so far. Please note that in recent KonText and
// Klogproc versions this is rather a fallback/offline functionality.

package alarm

import "log"

// BatchProcAlarm is a pseudo-alarm for batch processing which just
// logs information about total number of logged errors during processing.
type BatchProcAlarm struct {
	numErr int
}

func (bpa *BatchProcAlarm) OnError() {
	bpa.numErr++
}

func (bpa *BatchProcAlarm) Evaluate() {
	log.Printf("INFO: number of logged errors: %d", bpa.numErr)
}

func (bpa *BatchProcAlarm) Reset() {
	bpa.numErr = 0
}
