// Copyright 2018 Tomas Machalek <tomas.machalek@gmail.com>
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

package main

var helpTexts = []string{
	`Parse data provided either as a directory containing one or more application
log files or a Redis list (preferred). A proper JSON configuration file must be specified

{
    "logRedis": {
        "address": "127.0.0.1:6379",
        "database": 1,
        "queueKey": "kontext_log_queue"
    },
    "elasticServer": "http://elastic:9200",
    "elasticIndex": "trost_kontext",
    "appType": "kontext",
    "elasticSearchChunkSize": 1000,
    "elasticPushChunkSize": 2000,
    "elasticScrollTtl": "3m",
    "geoIPDbPath": "/path/to/GeoLite2-City.mmdb",
    "localTimezone": "+01:00",
    "anonymousUsers": [4230]
}
`,
	`Update each matching (defined by filter in "updates") record using a provided object (defined in "updateData"). NOTE: This is experimental.`,
}
