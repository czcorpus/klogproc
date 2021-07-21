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

package wag06

import (
	"log"
	"strings"

	"github.com/czcorpus/klogproc/load/accesslog"
)

const (
	actionSearch           = "search"
	actionTranslate        = "translate"
	actionCompare          = "compare"
	actionWordForms        = "word-forms"
	actionSimilarFreqWords = "similar-freq-words"
	actionSetTheme         = "set-theme"
	actionTelemetry        = "telemetry"
	actionSetUILang        = "set-ui-lang"
	actionGetLemmas        = "get-lemmas"
	actionEmbedded         = "embedded"
)

var (
	pathPrefixes = []string{"slovo-v-kostce", "word-at-a-glance", "wag", "wag-beta", "wdg", "wdglance"}
)

type actionArgs struct {
	action  string
	queries []string
	lang1   string
	lang2   string
}

func getAction(path string) actionArgs {
	ans := actionArgs{}
	items := strings.Split(strings.Trim(path, "/"), "/")
	if len(items) == 0 {
		return ans
	}
	var action string
	for _, prefix := range pathPrefixes {
		if len(items) > 1 && items[0] == prefix {
			action = items[1]
			items = items[2:]
		}
	}
	if action == "" {
		action = items[0]
		items = items[1:]
	}
	if !isProcessable(action) {
		return ans
	}
	ans.action = action
	switch ans.action {
	case actionSearch:
		if len(items) >= 2 {
			ans.lang1 = items[0]
			ans.queries = []string{items[1]}

		} else {
			log.Print("WARNING: ignoring legacy search action: ", path)
		}
	case actionTranslate:
		langItems := strings.Split(items[0], "--")
		ans.lang1 = langItems[0]
		if len(langItems) > 1 {
			ans.lang2 = langItems[1]

		} else {
			log.Print("WARNING: missing second lang info: ", path)
		}
		if len(items) > 1 {
			ans.queries = []string{items[1]}

		} else {
			log.Print("WARNING: missing query: ", path)
		}
	case actionCompare:
		if len(items) > 1 {
			ans.lang1 = items[0]
			if len(items) > 1 {
				ans.queries = strings.Split(items[1], "--")

			} else {
				log.Print("WARNING: missing query: ", path)
			}
		}
	}
	return ans
}

func isProcessable(action string) bool {
	return action == actionSearch || action == actionTranslate ||
		action == actionCompare || action == actionWordForms ||
		action == actionSimilarFreqWords || action == actionSetTheme ||
		action == actionTelemetry || action == actionSetUILang ||
		action == actionGetLemmas || action == actionEmbedded
}

// LineParser is a parser for reading KonText application logs
type LineParser struct {
	parser accesslog.LineParser
}

// ParseLine parses a HTTP access log format line
func (lp *LineParser) ParseLine(s string, lineNum int64) (*InputRecord, error) {
	parsed, err := lp.parser.ParseLine(s, lineNum)
	if err != nil {
		return &InputRecord{isProcessable: false}, err
	}
	action := getAction(parsed.Path)
	if action.action == "" {
		return &InputRecord{isProcessable: false}, nil
	}

	lcUA := strings.ToLower(parsed.UserAgent)

	ans := &InputRecord{
		isProcessable: true,
		Action:        action.action,
		Datetime:      parsed.Datetime,
		Request: Request{
			HTTPUserAgent:  parsed.UserAgent,
			HTTPRemoteAddr: parsed.IPAddress,
			RemoteAddr:     parsed.IPAddress, // TODO the same stuff as above?
			Referer:        parsed.Referrer,
		},
		ProcTime:            parsed.ProcTime,
		QueryType:           action.action, // for legacy reasons (otherwise it is redundant)
		Lang1:               action.lang1,
		Lang2:               action.lang2,
		Queries:             action.queries,
		HasPosSpecification: parsed.URLArgs.Get("pos") != "" || parsed.URLArgs.Get("lemma") != "",
		IsMobileClient:      strings.Contains(lcUA, "mobile") || strings.Contains(lcUA, "iphone") || strings.Contains(lcUA, "tablet") || strings.Contains(lcUA, "android"),
	}
	return ans, nil
}
