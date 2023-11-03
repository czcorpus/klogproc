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

package notifications

import (
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
)

const (
	defaultSender = "klogproc@localhost"
)

// Notifier is a general type representing a service
// for sending warnings to administrators
// (typically when a suspicious activity is detected from processed logs)
type Notifier interface {
	SendNotification(subject string, metadata map[string]any, paragraphs ...string) error
}

// defaultEmailNotifier provides basic e-mail notification
// as used by other parts of klogproc (e.g. sending alarm info).
// It should be instantiated via NewEmailNotifier.
type defaultEmailNotifier struct {
	conf *mail.NotificationConf
	loc  *time.Location
}

func (den *defaultEmailNotifier) SendNotification(subject string, metadata map[string]any, divContents ...string) error {
	return mail.SendNotification(den.conf, den.loc, mail.FormattedNotification{
		Subject: subject,
		Divs:    divContents,
	})
}
