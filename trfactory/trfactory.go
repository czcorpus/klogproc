// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
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

	"klogproc/email"
	"klogproc/load"
	"klogproc/servicelog"
	"klogproc/servicelog/apiguard"
	"klogproc/servicelog/kontext013"
	"klogproc/servicelog/kontext015"
	"klogproc/servicelog/kontext018"
	"klogproc/servicelog/korpusdb"
	"klogproc/servicelog/kwords"
	"klogproc/servicelog/mapka"
	"klogproc/servicelog/mapka2"
	"klogproc/servicelog/mapka3"
	"klogproc/servicelog/masm"
	"klogproc/servicelog/morfio"
	"klogproc/servicelog/shiny"
	"klogproc/servicelog/ske"
	"klogproc/servicelog/syd"
	"klogproc/servicelog/treq"
	"klogproc/servicelog/wag06"
	"klogproc/servicelog/wag07"
	"klogproc/servicelog/wsserver"
	"klogproc/users"
)

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(
	appType string,
	version string,
	bufferConf *load.BufferConf,
	userMap *users.UserMap,
	realtimeClock bool,
	emailNotifier email.MailNotifier,
) (servicelog.LogItemTransformer, error) {

	switch appType {
	case servicelog.AppTypeAPIGuard:
		return &apiguardTransformer{
			t: &apiguard.Transformer{},
		}, nil
	case servicelog.AppTypeAkalex, servicelog.AppTypeCalc, servicelog.AppTypeLists,
		servicelog.AppTypeQuitaUp, servicelog.AppTypeGramatikat:
		return &shinyTransformer{t: shiny.NewTransformer()}, nil
	case servicelog.AppTypeKontext, servicelog.AppTypeKontextAPI:
		switch version {
		case "0.13", "0.14":
			return &konText013Transformer{t: &kontext013.Transformer{}}, nil
		case "0.15", "0.16", "0.17":
			return &konText015Transformer{t: &kontext015.Transformer{}}, nil
		case "0.18":
			return &konText018Transformer{
				t: kontext018.NewTransformer(
					bufferConf,
					realtimeClock,
					emailNotifier,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case servicelog.AppTypeKwords:
		return &kwordsTransformer{t: &kwords.Transformer{}}, nil
	case servicelog.AppTypeKorpusDB:
		return &korpusDBTransformer{t: &korpusdb.Transformer{}}, nil
	case servicelog.AppTypeMapka:
		switch version {
		case "1":
			return &mapkaTransformer{t: mapka.NewTransformer()}, nil
		case "2":
			return &mapka2Transformer{t: mapka2.NewTransformer()}, nil
		case "3":
			return &mapka3Transformer{
				t: mapka3.NewTransformer(
					bufferConf,
					realtimeClock,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case servicelog.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{}}, nil
	case servicelog.AppTypeSke:
		return &skeTransformer{t: ske.NewTransformer(userMap)}, nil
	case servicelog.AppTypeSyd:
		return &sydTransformer{t: syd.NewTransformer(version)}, nil
	case servicelog.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{}}, nil
	case servicelog.AppTypeWag:
		switch version {
		case "0.6":
			return &wag06Transformer{t: &wag06.Transformer{}}, nil
		case "0.7":
			return &wag07Transformer{
				t: wag07.NewTransformer(
					bufferConf,
					realtimeClock,
					emailNotifier,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case servicelog.AppTypeWsserver:
		return &wsserverTransformer{t: &wsserver.Transformer{}}, nil
	case servicelog.AppTypeMasm:
		return &masmTransformer{t: &masm.Transformer{}}, nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
