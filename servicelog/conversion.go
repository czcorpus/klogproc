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

package servicelog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// TimezoneToInt returns number of minutes to add/subtract to apply
// to UTC to get actual local time reprezented by 'tz'.
func TimezoneToInt(tz string) (int, error) {
	sgn := 1
	if tz[0] == '-' {
		sgn = -1

	} else if tz[0] != '+' {
		return 0, fmt.Errorf("cannot parse %s as timezone value", tz)
	}
	items := strings.Split(tz[1:], ":")
	if len(items) != 2 {
		return 0, fmt.Errorf("cannot parse %s as timezone value", tz)
	}
	v1, err := strconv.Atoi(items[0])
	if err != nil {
		return 0, err
	}
	v2, err := strconv.Atoi(items[1])
	if err != nil {
		return 0, err
	}
	return sgn * (60*v1 + v2), nil
}

// ImportBool imports typical bool formats (as supported by Go) with
// additional support for an empty space, 'yes' and 'no' strings.
func ImportBool(v, keyName string) (bool, error) {
	if v == "" {
		return false, nil
	}
	if v == "yes" {
		return true, nil
	}
	if v == "no" {
		return false, nil
	}
	ans, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid data for %s: %s", keyName, v)
	}
	return ans, nil
}

// ConvertDatetimeString imports ISO 8601 datetime string. In case
// of a parsing error, "zero" time instance is created.
func ConvertDatetimeString(datetime string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", datetime)
	if err == nil {
		return t
	}
	log.Error().Err(err).Str("value", datetime).Msgf("failed to convert datetime string")
	return time.Time{}
}

func ConvertDatetimeStringWithMillis(datetime string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000000-07:00", datetime)
	if err == nil {
		return t
	}
	log.Error().Err(err).Str("value", datetime).Msgf("failed to convert datetime string (with millis)")
	return time.Time{}
}

func ConvertDatetimeStringNoTZ(datetime string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", datetime)
	if err == nil {
		return t
	}
	log.Error().Err(err).Str("value", datetime).Msgf("failed to convert datetime string (no tz)")
	return time.Time{}
}

func ConvertDatetimeStringWithMillisNoTZ(datetime string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000000", datetime)
	if err == nil {
		return t
	}
	log.Error().Err(err).Str("value", datetime).Msgf("failed to convert datetime string (with millis, no tz)")
	return time.Time{}
}

func ConvertAccessLogDatetimeString(datetime string) time.Time {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", datetime)
	if err == nil {
		return t
	}
	log.Error().Err(err).Str("value", datetime).Msgf("failed to convert datetime string (access log format)")
	return time.Time{}
}
