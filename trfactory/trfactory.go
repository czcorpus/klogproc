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

	"klogproc/notifications"
	"klogproc/servicelog"
	"klogproc/servicelog/apiguard"
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
)

// GetStaticLogTransformer returns a type-safe transformer for a concrete app type
func GetStaticLogTransformer(
	logConf servicelog.LogProcConf,
	anonymousUsers []int,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
) (servicelog.LogItemTransformer, error) {

	appType := logConf.GetAppType()
	version := logConf.GetVersion()
	bufferConf := logConf.GetBuffer()

	switch appType {
	case servicelog.AppTypeAPIGuard:
		return &apiguard.Transformer{}, nil
	case servicelog.AppTypeAkalex, servicelog.AppTypeCalc, servicelog.AppTypeLists,
		servicelog.AppTypeQuitaUp, servicelog.AppTypeGramatikat:
		return shiny.NewTransformer(appType, anonymousUsers), nil
	case servicelog.AppTypeKontext:
		switch version {
		case servicelog.AppVersionKontext013, servicelog.AppVersionKontext014:
			return &kontext013.Transformer{
				AnonymousUsers: anonymousUsers}, nil
		case servicelog.AppVersionKontext015,
			servicelog.AppVersionKontext016,
			servicelog.AppVersionKontext017:
			return &kontext015.Transformer{
				AnonymousUsers: anonymousUsers,
				IsAPI:          true,
			}, nil
		case servicelog.AppVersionKontext017API:
			return &kontext015.Transformer{AnonymousUsers: anonymousUsers}, nil
		case servicelog.AppVersionKontext018:
			return kontext018.NewTransformer(
				bufferConf,
				realtimeClock,
				emailNotifier,
				anonymousUsers,
			), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case servicelog.AppTypeKwords:
		switch version {
		case servicelog.AppVersionKwords1:
			return &kwords.Transformer{AnonymousUsers: anonymousUsers}, nil
		case servicelog.AppVersionKwords2:
			return &kwords2.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KWords version: %s", version)
		}

	case servicelog.AppTypeKorpusDB:
		return korpusdb.NewTransformer(), nil
	case servicelog.AppTypeMapka:
		switch version {
		case servicelog.AppVersionMapka1:
			return mapka.NewTransformer(anonymousUsers), nil
		case servicelog.AppVersionMapka2:
			return mapka2.NewTransformer(anonymousUsers), nil
		case servicelog.AppVersionMapka3:
			return mapka3.NewTransformer(bufferConf, anonymousUsers, realtimeClock), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case servicelog.AppTypeMorfio:
		return &morfio.Transformer{
			AnonymousUsers: anonymousUsers}, nil
	case servicelog.AppTypeSke:
		return ske.NewTransformer(anonymousUsers), nil
	case servicelog.AppTypeSyd:
		return syd.NewTransformer(version, anonymousUsers), nil
	case servicelog.AppTypeTreq:
		switch version {
		case servicelog.AppVersionTreq1API:
			return &treqapi.Transformer{AnonymousUsers: anonymousUsers}, nil
		default:
			return &treq.Transformer{AnonymousUsers: anonymousUsers}, nil
		}
	case servicelog.AppTypeWag:
		switch version {
		case servicelog.AppVersionWag06:
			return &wag06.Transformer{}, nil
		case servicelog.AppVersionWag07:
			return wag07.NewTransformer(bufferConf, anonymousUsers, realtimeClock, emailNotifier), nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case servicelog.AppTypeWsserver:
		return &wsserver.Transformer{}, nil
	case servicelog.AppTypeMasm:
		return &masm.Transformer{}, nil
	case servicelog.AppTypeMquery:
		return &mquery.Transformer{}, nil
	case servicelog.AppTypeMquerySRU:
		return &mquerysru.Transformer{}, nil
	case servicelog.AppTypeVLO:
		return &vlo.Transformer{}, nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
