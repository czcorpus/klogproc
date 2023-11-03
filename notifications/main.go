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
	"errors"
	"fmt"
	goMail "net/mail"
	"strings"
	"time"

	"github.com/czcorpus/cnc-gokit/mail"
	"github.com/czcorpus/conomi/client"
	"github.com/rs/zerolog/log"
)

// NewNotifier is a factory function for e-mail/Conomi notification.
// Because of a difference between both configurations, the Klogproc
// config type contains two sections - here represented by `conf` and
// `conf2`. Both configs are mutually exclusive and in case both
// are provided, the function returns and error.
//
// Missing sender is replaced by a default value.
func NewNotifier(
	conf *mail.NotificationConf,
	conf2 *client.ConomiClientConf,
	loc *time.Location,
) (Notifier, error) {
	if conf != nil && conf2 != nil {
		return nil, errors.New("either Conomi or e-mail notifier can be configured")
	}
	if conf2 != nil {
		cclient := client.NewConomiClient(*conf2)
		return &conomiNotifier{conf: conf2, client: cclient}, nil

	} else if conf != nil {
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
		log.Info().Msgf(
			"creating e-mail sender with recipient(s) %s", strings.Join(conf.Recipients, ", "))
		return &defaultEmailNotifier{conf: conf, loc: loc}, nil
	}
	return &nullEmailNotifier{}, nil

}
