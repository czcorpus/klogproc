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

import "github.com/rs/zerolog/log"

// nullEmailNotifier is a regular mail sender replacement
// in case mailing is not configured
type nullEmailNotifier struct {
}

func (den *nullEmailNotifier) SendNotification(subject string, metadata map[string]any, paragraphs ...string) error {
	log.Warn().
		Str("subject", subject).
		Strs("body", paragraphs).
		Msg("not sending e-mail notification - not configured")
	return nil
}
