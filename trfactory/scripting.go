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

	"github.com/czcorpus/klogproc-core/analysis"
	"github.com/czcorpus/klogproc-core/scripting"
	"github.com/czcorpus/klogproc-core/storage"
	k015Core "github.com/czcorpus/klogproc-core/storage/kontext015"
	kdbCore "github.com/czcorpus/klogproc-core/storage/korpusdb"
	kwords2Core "github.com/czcorpus/klogproc-core/storage/kwords2"
	skeCore "github.com/czcorpus/klogproc-core/storage/ske"
	treqCore "github.com/czcorpus/klogproc-core/storage/treq"
)

// GetLogTransformer creates a log transformer with optional support for Lua scripting.
// In case there is no script defined or the application type does not support scripting
// access, the transformer delegates its methods to the traditional "static" transformer
// (i.e. the one compiled directly to klogproc).
func GetLogTransformer(
	logConf storage.LogProcConf,
	anonymousUsers []int,
	realtimeClock bool,
	emailNotifier analysis.Notifier,
) (*scripting.Transformer, error) {
	tr, err := GetStaticLogTransformer(logConf, anonymousUsers, realtimeClock, emailNotifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
	}

	if logConf.GetScriptPath() == "" {
		return scripting.NewTransformer(nil, tr), nil
	}

	version := logConf.GetVersion()
	switch logConf.GetAppType() {
	case storage.AppTypeKontext:
		if version == storage.AppVersionKontext018 ||
			version == storage.AppVersionKontext017 ||
			version == storage.AppVersionKontext017API ||
			version == storage.AppVersionKontext016 ||
			version == storage.AppVersionKontext015 {
			env, err := scripting.CreateEnvironment(
				logConf,
				anonymousUsers,
				tr,
				func() storage.OutputRecord { return &k015Core.OutputRecord{} },
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
			}
			return scripting.NewTransformer(env, tr), nil
		}
	case storage.AppTypeKorpusDB:
		env, err := scripting.CreateEnvironment(
			logConf,
			anonymousUsers,
			tr,
			func() storage.OutputRecord { return &kdbCore.OutputRecord{} },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
		}
		return scripting.NewTransformer(env, tr), nil
	case storage.AppTypeKwords:
		if version == storage.AppVersionKwords2 {
			env, err := scripting.CreateEnvironment(
				logConf,
				anonymousUsers,
				tr,
				func() storage.OutputRecord { return &kwords2Core.OutputRecord{} },
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
			}
			return scripting.NewTransformer(env, tr), nil
		}
	case storage.AppTypeSke:
		env, err := scripting.CreateEnvironment(
			logConf,
			anonymousUsers,
			tr,
			func() storage.OutputRecord { return &skeCore.OutputRecord{} },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
		}
		return scripting.NewTransformer(env, tr), nil
	case storage.AppTypeTreq:
		env, err := scripting.CreateEnvironment(
			logConf,
			anonymousUsers,
			tr,
			func() storage.OutputRecord { return &treqCore.OutputRecord{} },
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create scripting transformer for %s: %w", logConf.GetAppType(), err)
		}
		return scripting.NewTransformer(env, tr), nil

	}

	return nil, fmt.Errorf("klogproc does not support scripting for %s logs", logConf.GetAppType())
}
