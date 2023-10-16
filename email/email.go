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

package email

import (
	"fmt"
	goMail "net/mail"
	"strings"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/rs/zerolog/log"
)

const (
	defaultSender = "klogproc@localhost"
)

// MailNotifier is a general type allowing sending messages to
// a predefined list of recipients
type MailNotifier interface {
	SendNotification(subject string, message ...string) error
	SendFormattedNotification(subject string, divContents ...string) error
}

// nullEmailNotifier is a regular mail sender replacement
// in case mailing is not configured
type nullEmailNotifier struct {
}

func (den *nullEmailNotifier) SendNotification(subject string, message ...string) error {
	log.Warn().
		Str("subject", subject).
		Strs("body", message).
		Msg("not sending e-mail notification - not configured")
	return nil
}

func (den *nullEmailNotifier) SendFormattedNotification(subject string, divContents ...string) error {
	log.Warn().
		Str("subject", subject).
		Strs("body", divContents).
		Msg("not sending e-mail notification - not configured")
	return nil
}

// defaultEmailNotifier provides basic e-mail notification
// as used by other parts of klogproc (e.g. sending alarm info).
// It should be instantiated via NewEmailNotifier.
type defaultEmailNotifier struct {
	conf *mail.NotificationConf
	loc  *time.Location
}

// SendNotification sends a general e-mail notification based on
// a respective monitoring configuration. The 'alarmToken' argument
// can be nil - in such case the 'turn of the alarm' text won't be
// part of the message.
func (den *defaultEmailNotifier) SendNotification(subject string, message ...string) error {
	return mail.SendNotification(den.conf, den.loc, mail.Notification{
		Subject:    subject,
		Paragraphs: message,
	})
}

func (den *defaultEmailNotifier) SendFormattedNotification(subject string, divContents ...string) error {
	return mail.SendNotification(den.conf, den.loc, mail.FormattedNotification{
		Subject: subject,
		Divs:    divContents,
	})
}

// NewEmailNotifier is a factory function for email notification
// In case of missing configuration, the function returns an error.
// Missing sender is replaced by a default value.
func NewEmailNotifier(
	conf *mail.NotificationConf,
	loc *time.Location,
) (MailNotifier, error) {
	if conf == nil {
		return &nullEmailNotifier{}, nil
	}
	if conf.Sender == "" {
		log.Warn().Msgf("e-mail sender not set - using default %s", defaultSender)
		conf.Sender = defaultSender
	}
	validated := append([]string{conf.Sender}, conf.Recipients...)
	for _, addr := range validated {
		if _, err := goMail.ParseAddress(addr); err != nil {
			return nil, fmt.Errorf("incorrect e-mail address %s: %s", addr, err)
		}
	}
	log.Info().Msgf("creating e-mail sender with recipient(s) %s", strings.Join(conf.Recipients, ", "))
	return &defaultEmailNotifier{conf: conf, loc: loc}, nil
}
