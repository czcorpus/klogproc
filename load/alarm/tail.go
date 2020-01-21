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

import (
	"fmt"
	"log"
	"time"
)

// TailProcAlarm counts number of logged errors and if the total
// number during a defined time interval reaches a defined size,
// e-mail notification is triggered.
type TailProcAlarm struct {
	errCountTimeRangeSecs int
	notifier              mailNotifier
	lastErrors            []int64
	errIdx                int
	fileInfo              tailFileDescriber
}

// OnError inserts timestamp of the error detection event.
func (tpa *TailProcAlarm) OnError() {
	tpa.errIdx = (tpa.errIdx + 1) % len(tpa.lastErrors)
	tpa.lastErrors[tpa.errIdx] = time.Now().Unix()
}

// Evaluate looks for oldest and newest errors and if all
// the internal slots are full and the interval is smaller
// or equal of a defined value, an alarm e-mail is sent.
func (tpa *TailProcAlarm) Evaluate() {
	oldest, newest := findRange(tpa.lastErrors)
	if oldest > 0 && newest-oldest <= int64(tpa.errCountTimeRangeSecs) {
		msg := fmt.Sprintf("Too many errors (%d) logged within file %s during defined interval of %d seconds",
			len(tpa.lastErrors), tpa.fileInfo.GetPath(), tpa.errCountTimeRangeSecs)
		subj := fmt.Sprintf("Klogproc ERROR alarm for file %s (type %s)", tpa.fileInfo.GetPath(),
			tpa.fileInfo.GetAppType())
		log.Printf("INFO: sending alarm notification for %s", tpa.fileInfo.GetPath())
		err := tpa.notifier.SendNotification(subj, msg)
		if err != nil {
			log.Print("ERROR: ", err)
		}
		tpa.Reset()
	}
}

// Reset clears the whole state of the alarm.
func (tpa *TailProcAlarm) Reset() {
	for i := range tpa.lastErrors {
		tpa.lastErrors[i] = 0
	}
	tpa.errIdx = 1
}

// NewTailProcAlarm is a recommended factory for TailProcAlarm type
func NewTailProcAlarm(maxNumErr int, errCountTimeRangeSecs int, fileInfo tailFileDescriber, notifier mailNotifier) *TailProcAlarm {
	return &TailProcAlarm{
		notifier:              notifier,
		errCountTimeRangeSecs: errCountTimeRangeSecs,
		lastErrors:            make([]int64, maxNumErr),
		errIdx:                1, // we want the interval to be super-long until all the slots in lastErrors are filled in
		fileInfo:              fileInfo,
	}
}
