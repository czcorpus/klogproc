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

	"klogproc/servicelog/apiguard"
	apiguardKontext018 "klogproc/servicelog/apiguard-kontext018"
	apiguardKwords "klogproc/servicelog/apiguard-kwords"
	apiguardMquery "klogproc/servicelog/apiguard-mquery"
	apiguardTreq "klogproc/servicelog/apiguard-treq"
	"klogproc/servicelog/kontext013"
	"klogproc/servicelog/kontext015"
	"klogproc/servicelog/kontext018"
	"klogproc/servicelog/korpusdb"
	"klogproc/servicelog/kwords"
	"klogproc/servicelog/kwords2"
	"klogproc/servicelog/mapka"
	"klogproc/servicelog/mapka2"
	"klogproc/servicelog/mapka3"
	"klogproc/servicelog/masm"
	"klogproc/servicelog/morfio"
	"klogproc/servicelog/mquery"
	"klogproc/servicelog/mquerysru"
	"klogproc/servicelog/shiny"
	"klogproc/servicelog/ske"
	"klogproc/servicelog/syd"
	"klogproc/servicelog/treq"
	"klogproc/servicelog/treqapi"
	"klogproc/servicelog/vlo"
	"klogproc/servicelog/wag06"
	"klogproc/servicelog/wag07"
	"klogproc/servicelog/wsserver"

	"github.com/czcorpus/klogproc-core/analysis"
	"github.com/czcorpus/klogproc-core/storage"
)

// GetStaticLogTransformer returns a type-safe transformer for a concrete app type
func GetStaticLogTransformer(
	logConf storage.LogProcConf,
	anonymousUsers []int,
	realtimeClock bool,
	emailNotifier analysis.Notifier,
) (storage.LogItemTransformer, error) {

	appType := logConf.GetAppType()
	version := logConf.GetVersion()
	bufferConf := logConf.GetBuffer()

	switch appType {
	case storage.AppTypeAPIGuard:
		return &apiguard.Transformer{}, nil
	case storage.AppTypeAPIGuardMquery:
		return &apiguardMquery.Transformer{}, nil
	case storage.AppTypeAPIGuardKontext:
		switch version {
		case storage.AppVersionKontext018:
			return &apiguardKontext018.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return nil, fmt.Errorf("cannot create ApiGuard transformer, unsupported KonText version: %s", version)
		}
	case storage.AppTypeAPIGuardTreq:
		switch version {
		case storage.AppVersionTreq1API:
			return nil, fmt.Errorf("cannot create ApiGuard transformer, unsupported Treq version: %s", version)
		default:
			return &apiguardTreq.Transformer{AnonymousUsers: anonymousUsers}, nil
		}
	case storage.AppTypeAPIGuardKwords:
		switch version {
		case storage.AppVersionKwords1:
			return &apiguardKwords.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return nil, fmt.Errorf("cannot create ApiGuard transformer, unsupported KWords version: %s", version)
		}
	case storage.AppTypeAkalex, storage.AppTypeCalc, storage.AppTypeLists,
		storage.AppTypeQuitaUp, storage.AppTypeGramatikat:
		return shiny.NewTransformer(appType, anonymousUsers), nil
	case storage.AppTypeKontext:
		switch version {
		case storage.AppVersionKontext013, storage.AppVersionKontext014:
			return &kontext013.Transformer{
				AnonymousUsers: anonymousUsers}, nil
		case storage.AppVersionKontext015,
			storage.AppVersionKontext016,
			storage.AppVersionKontext017:
			return &kontext015.Transformer{
				AnonymousUsers: anonymousUsers,
				IsAPI:          true,
			}, nil
		case storage.AppVersionKontext017API:
			return &kontext015.Transformer{AnonymousUsers: anonymousUsers}, nil
		case storage.AppVersionKontext018:
			return kontext018.NewTransformer(
				bufferConf,
				realtimeClock,
				emailNotifier,
				anonymousUsers,
			), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case storage.AppTypeKwords:
		switch version {
		case storage.AppVersionKwords1:
			return &kwords.Transformer{AnonymousUsers: anonymousUsers}, nil
		case storage.AppVersionKwords2:
			return &kwords2.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KWords version: %s", version)
		}

	case storage.AppTypeKorpusDB:
		return korpusdb.NewTransformer(), nil
	case storage.AppTypeMapka:
		switch version {
		case storage.AppVersionMapka1:
			return mapka.NewTransformer(anonymousUsers), nil
		case storage.AppVersionMapka2:
			return mapka2.NewTransformer(anonymousUsers), nil
		case storage.AppVersionMapka3:
			return mapka3.NewTransformer(bufferConf, anonymousUsers, realtimeClock), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case storage.AppTypeMorfio:
		return &morfio.Transformer{
			AnonymousUsers: anonymousUsers}, nil
	case storage.AppTypeSke:
		return ske.NewTransformer(anonymousUsers), nil
	case storage.AppTypeSyd:
		return syd.NewTransformer(version, anonymousUsers), nil
	case storage.AppTypeTreq:
		switch version {
		case storage.AppVersionTreq1API:
			return &treqapi.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return &treq.Transformer{AnonymousUsers: anonymousUsers}, nil
		}
	case storage.AppTypeWag:
		switch version {
		case storage.AppVersionWag06:
			return &wag06.Transformer{}, nil
		case storage.AppVersionWag07:
			return wag07.NewTransformer(bufferConf, anonymousUsers, realtimeClock, emailNotifier), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case storage.AppTypeWsserver:
		return &wsserver.Transformer{}, nil
	case storage.AppTypeMasm:
		return &masm.Transformer{}, nil
	case storage.AppTypeMquery:
		return &mquery.Transformer{}, nil
	case storage.AppTypeMquerySRU:
		return &mquerysru.Transformer{}, nil
	case storage.AppTypeVLO:
		return &vlo.Transformer{}, nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
