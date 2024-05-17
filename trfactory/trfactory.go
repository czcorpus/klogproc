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

	"klogproc/load"
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
	"klogproc/servicelog/vlo"
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
	excludeIpList servicelog.ExcludeIPList,
	realtimeClock bool,
	emailNotifier notifications.Notifier,
) (servicelog.LogItemTransformer, error) {

	switch appType {
	case servicelog.AppTypeAPIGuard:
		return &apiguardTransformer{
			t: &apiguard.Transformer{
				ExcludeIPList: excludeIpList,
			},
		}, nil
	case servicelog.AppTypeAkalex, servicelog.AppTypeCalc, servicelog.AppTypeLists,
		servicelog.AppTypeQuitaUp, servicelog.AppTypeGramatikat:
		return &shinyTransformer{
			t: shiny.NewTransformer(excludeIpList),
		}, nil
	case servicelog.AppTypeKontext, servicelog.AppTypeKontextAPI:
		switch version {
		case "0.13", "0.14":
			return &konText013Transformer{
				t: &kontext013.Transformer{
					ExcludeIPList: excludeIpList,
				},
			}, nil
		case "0.15", "0.16", "0.17":
			return &konText015Transformer{
				t: &kontext015.Transformer{
					ExcludeIPList: excludeIpList,
				},
			}, nil
		case "0.18":
			return &konText018Transformer{
				t: kontext018.NewTransformer(
					bufferConf,
					realtimeClock,
					emailNotifier,
					excludeIpList,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KonText version: %s", version)
		}
	case servicelog.AppTypeKwords:
		switch version {
		case "1":
			return &kwordsTransformer{
				t: &kwords.Transformer{
					ExcludeIPList: excludeIpList,
				},
			}, nil
		case "2":
			return &kwords2Transformer{
				t: &kwords2.Transformer{
					ExcludeIPList: excludeIpList,
				}}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported KWords version: %s", version)
		}

	case servicelog.AppTypeKorpusDB:
		return &korpusDBTransformer{t: &korpusdb.Transformer{
			ExcludeIPList: excludeIpList,
		}}, nil
	case servicelog.AppTypeMapka:
		switch version {
		case "1":
			return &mapkaTransformer{
				t: mapka.NewTransformer(excludeIpList),
			}, nil
		case "2":
			return &mapka2Transformer{
				t: mapka2.NewTransformer(excludeIpList),
			}, nil
		case "3":
			return &mapka3Transformer{
				t: mapka3.NewTransformer(
					bufferConf,
					excludeIpList,
					realtimeClock,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported Mapka version: %s", version)
		}
	case servicelog.AppTypeMorfio:
		return &morfioTransformer{t: &morfio.Transformer{
			ExcludeIPList: excludeIpList,
		}}, nil
	case servicelog.AppTypeSke:
		return &skeTransformer{
				t: ske.NewTransformer(userMap, excludeIpList),
			},
			nil
	case servicelog.AppTypeSyd:
		return &sydTransformer{
			t: syd.NewTransformer(version, excludeIpList),
		}, nil
	case servicelog.AppTypeTreq:
		return &treqTransformer{t: &treq.Transformer{
			ExcludeIPList: excludeIpList,
		}}, nil
	case servicelog.AppTypeWag:
		switch version {
		case "0.6":
			return &wag06Transformer{
				t: &wag06.Transformer{
					ExcludeIPList: excludeIpList,
				},
			}, nil
		case "0.7":
			return &wag07Transformer{
				t: wag07.NewTransformer(
					bufferConf,
					excludeIpList,
					realtimeClock,
					emailNotifier,
				),
			}, nil
		default:
			return nil, fmt.Errorf("cannot create transformer, unsupported WaG version: %s", version)
		}
	case servicelog.AppTypeWsserver:
		return &wsserverTransformer{
			t: &wsserver.Transformer{
				ExcludeIPList: excludeIpList,
			},
		}, nil
	case servicelog.AppTypeMasm:
		return &masmTransformer{t: &masm.Transformer{
			ExcludeIPList: excludeIpList,
		}}, nil
	case servicelog.AppTypeMquery:
		return &mqueryTransformer{t: &mquery.Transformer{
			ExcludeIPList: excludeIpList,
		}}, nil
	case servicelog.AppTypeMquerySRU:
		return &mquerySRUTransformer{
				t: &mquerysru.Transformer{
					ExcludeIPList: excludeIpList,
				},
			},
			nil
	case servicelog.AppTypeVLO:
		return &VLOTransformer{
				t: &vlo.Transformer{
					ExcludeIPList: excludeIpList,
				},
			},
			nil
	default:
		return nil, fmt.Errorf("cannot find log transformer for app type %s", appType)
	}
}
