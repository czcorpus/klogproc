// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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

package trfactory

import (
	"fmt"
	"klogproc/notifications"
	"klogproc/scripting"
	"klogproc/servicelog"
	"klogproc/servicelog/kontext018"
)

// GetLogTransformer creates a log transformer with optional support for Lua scripting.
// In case there is no script defined or the application type does not support scripting
// access, the transformer delegates its methods to the traditional "static" transformer
// (i.e. the one compiled directly to klogproc).
func GetLogTransformer(
	logConf servicelog.LogProcConf,
	anonymousUsers []int,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
) (*scripting.Transformer, error) {
	tr, err := GetStaticLogTransformer(logConf, anonymousUsers, realtimeClock, emailNotifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
	}
	switch logConf.GetAppType() {
	case servicelog.AppTypeKontext, servicelog.AppTypeKontextAPI:
		if logConf.GetScriptPath() != "" {
			env, err := scripting.CreateEnvironment(
				logConf, tr, func() servicelog.OutputRecord { return &kontext018.OutputRecord{} })
			if err != nil {
				return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
			}
			return scripting.NewTransformer(env, tr), nil
		}
	}

	return scripting.NewTransformer(nil, tr), nil
}
