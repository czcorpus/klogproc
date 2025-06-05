// Copyright 2025 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2025 Institute of the Czech National Corpus,
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

package healthchk

import (
	"context"
	"fmt"
	"klogproc/notifications"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type updateInfo struct {
	logPath string
	dt      time.Time
}

type logInfo struct {
	lastDatetime time.Time
}

type ConomiNotifier struct {
	logs            map[string]logInfo
	ctx             context.Context
	ticker          *time.Ticker
	incomingUpdates chan updateInfo
	maxInactivity   map[string]time.Duration
	dataLock        sync.Mutex
	notifier        notifications.Notifier
	notificationTag string
}

// Update stores information about log's last change.
// The function can be called concurrently as it internally uses
// a channel to add new values
func (lwatch *ConomiNotifier) Ping(logPath string, dt time.Time) {
	lwatch.incomingUpdates <- updateInfo{logPath: logPath, dt: dt}
}

func (lwatch *ConomiNotifier) checkStatus() {
	lwatch.dataLock.Lock()
	defer lwatch.dataLock.Unlock()
	for logPath, v := range lwatch.logs {
		if time.Since(v.lastDatetime) > lwatch.maxInactivity[logPath] {
			go func() {
				subj := fmt.Sprintf(
					"log file %s seems inactive for too long (limit: %01.0f sec.)",
					logPath,
					lwatch.maxInactivity[logPath].Seconds(),
				)
				meta := map[string]any{}
				if lwatch.notifier != nil {
					err := lwatch.notifier.SendNotification(lwatch.notificationTag, subj, meta)
					if err != nil {
						log.Error().
							Err(err).
							// just so we can see this as there is no watchdog looking for this watchdog
							Str("severity", "=============== !!!!!!!!!! ==================").
							Msg("failed to send inactivity warning")
					}

				} else {
					log.Error().
						Str("subject", subj).
						Msg("====== log inactivity detected =======")
				}
			}()
		}
	}
}

func (lwatch *ConomiNotifier) goRegularCheck(intervalSecs int) {
	lwatch.ticker = time.NewTicker(time.Second * time.Duration(intervalSecs))
	go func() {
		for {
			select {
			case <-lwatch.ctx.Done():
				log.Warn().Msg("LogUpdateWatch closing due to cancellation")
				lwatch.ticker.Stop()
				return

			case <-lwatch.ticker.C:
				lwatch.checkStatus()
			case upd := <-lwatch.incomingUpdates:
				rec, ok := lwatch.logs[upd.logPath]
				if !ok {
					log.Error().
						Str("file", upd.logPath).
						Msg("LogUpdateWatch encountered an unconfigured file (probably a config. error)")
					rec = logInfo{}
				}
				rec.lastDatetime = upd.dt
				lwatch.logs[upd.logPath] = rec
			}
		}
	}()
}

// NewConomiNotifier
// note - the notifier can be nil in which case, only Klogproc's log will be used
// to put the alarms in.
func NewConomiNotifier(
	ctx context.Context,
	filesToWatch []string,
	tz *time.Location,
	intervalSecs int,
	maxInactivitySecs map[string]int,
	notifier notifications.Notifier,
	notificationTag string,
) *ConomiNotifier {

	logs := make(map[string]logInfo)
	for _, f := range filesToWatch {
		logs[f] = logInfo{lastDatetime: time.Now().In(tz)}
	}
	maxInactivity := make(map[string]time.Duration)
	for filePath, limitSecs := range maxInactivitySecs {
		maxInactivity[filePath] = time.Duration(limitSecs) * time.Second
	}
	ans := &ConomiNotifier{
		logs:            logs,
		ctx:             ctx,
		incomingUpdates: make(chan updateInfo, 1000),
		maxInactivity:   maxInactivity,
		notifier:        notifier,
		notificationTag: notificationTag,
	}
	ans.goRegularCheck(intervalSecs)
	return ans
}
