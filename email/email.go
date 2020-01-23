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
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/czcorpus/klogproc/config"
)

const (
	defaultSender = "klogproc@localhost"
)

// DefaultEmailNotifier provides basic e-mail notification
// as used by other parts of klogproc (e.g. sending alarm info).
// It should be instantiated via NewEmailNotifier.
type DefaultEmailNotifier struct {
	conf       *config.Email
	recipients []*mail.Address
}

// SendNotification sends a general e-mail notification based on
// a respective monitoring configuration. The 'alarmToken' argument
// can be nil - in such case the 'turn of the alarm' text won't be
// part of the message.
func (den *DefaultEmailNotifier) SendNotification(subject, message string) error {
	client, err := smtp.Dial(den.conf.SMTPServer)
	if err != nil {
		return err
	}
	defer client.Close()
	client.Mail(den.conf.Sender)
	for _, rcpt := range den.recipients {
		err = client.Rcpt(rcpt.Address)

		if err != nil {
			return err
		}
	}

	wc, err := client.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	headers := make(map[string]string)
	headers["From"] = den.conf.Sender
	headers["To"] = strings.Join(den.conf.NotificationEmails, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	body := ""
	for k, v := range headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "<p>" + message + "</p>\r\n\r\n"
	buf := bytes.NewBufferString(body)
	_, err = buf.WriteTo(wc)
	return err
}

// NewEmailNotifier is a factory function for email notification
// In case of missing configuration, the function returns an error.
// Missing sender is replaced by a default value.
func NewEmailNotifier(conf *config.Email) (*DefaultEmailNotifier, error) {
	if len(conf.NotificationEmails) == 0 || conf.SMTPServer == "" {
		return nil, errors.New("cannot create e-mail sender, missing configuration")
	}
	if conf.Sender == "" {
		log.Printf("WARNING: e-mail sender not set - using default %s", defaultSender)
		conf.Sender = defaultSender
	}
	recipients := make([]*mail.Address, len(conf.NotificationEmails))
	var err error
	for i, addr := range conf.NotificationEmails {
		recipients[i], err = mail.ParseAddress(addr)
		if err != nil {
			return nil, fmt.Errorf("address <%s> not parsed: %s", addr, err)
		}
	}
	log.Printf("INFO: creating e-mail sender with recipient(s) %s", strings.Join(conf.NotificationEmails, ", "))
	return &DefaultEmailNotifier{conf: conf, recipients: recipients}, nil
}
