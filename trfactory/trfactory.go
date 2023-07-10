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

	"klogproc/conversion"
	"klogproc/conversion/apiguard"
	"klogproc/conversion/kontext013"
	"klogproc/conversion/kontext015"
	"klogproc/conversion/kontext018"
	"klogproc/conversion/korpusdb"
	"klogproc/conversion/kwords"
	"klogproc/conversion/mapka"
	"klogproc/conversion/mapka2"
	"klogproc/conversion/mapka3"
	"klogproc/conversion/masm"
	"klogproc/conversion/morfio"
	"klogproc/conversion/shiny"
	"klogproc/conversion/ske"
	"klogproc/conversion/syd"
	"klogproc/conversion/treq"
	"klogproc/conversion/wag06"
	"klogproc/conversion/wag07"
	"klogproc/conversion/wsserver"
	"klogproc/users"
)

// GetLogTransformer returns a type-safe transformer for a concrete app type
func GetLogTransformer(
	appType string,
	version string,
	historyLookupSecs int,
	userMap *users.UserMap,
) (conversion.LogItemTransformer, error) {

	switch appType {
	case conversion.AppTypeAPIGuard:
		return &apiguardTransformer{
			t: &apiguard.Transformer{},
		}, nil
	case conversion.AppTypeAkalex, conversion.AppTypeCalc, conversion.AppTypeLists,
		conversion.AppTypeQuitaUp, conversion.AppTypeGramatikat:
		return &shinyTransformer{t: shiny.NewTransformer()}, nil
	case conversion.AppTypeKontext, conversion.AppTypeKontextAPI:
		switch version {
		case "0.13", "0.14":
			return &konText013Transformer{t: &kontext013.Transformer{}}, nil
		case "0.15", "0.16", "0.17":
			return &konText015Transformer{t: &kontext015.Transformer{}}, nil
		case "0.18":
			return &konText018Transformer{t: &kontext018.Transformer{}}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case conversion.AppTypeKwords:
		return &kwordsTransformer{t: &kwords.Transformer{}}, nil
	case conversion.AppTypeKorpusDB:
		return &korpusDBTransformer{t: &korpusdb.Transformer{}}, nil
	case conversion.AppTypeMapka:
		switch version {
		case "1":
			return &mapkaTransformer{t: mapka.NewTransformer()}, nil
		case "2":
			return &mapka2Transformer{t: mapka2.NewTransformer()}, nil
		case "3":
			return &mapka3Transformer{t: mapka3.NewTransformer(historyLookupSecs)}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case conversion.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{}}, nil
	case conversion.AppTypeSke:
		return &skeTransformer{t: ske.NewTransformer(userMap)}, nil
	case conversion.AppTypeSyd:
		return &sydTransformer{t: syd.NewTransformer(version)}, nil
	case conversion.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{}}, nil
	case conversion.AppTypeWag:
		switch version {
		case "0.6":
			return &wag06Transformer{t: &wag06.Transformer{}}, nil
		case "0.7":
			return &wag07Transformer{t: &wag07.Transformer{}}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case conversion.AppTypeWsserver:
		return &wsserverTransformer{t: &wsserver.Transformer{}}, nil
	case conversion.AppTypeMasm:
		return &masmTransformer{t: &masm.Transformer{}}, nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
