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

package ctype

import (
	"strings"

	"klogproc/conversion"
)

type LegacyClientTypeAnalyzer struct {
}

// AgentIsMonitor returns true if user agent information
// matches one of "bots" used by the Instatute Czech National Corpus
// to monitor service availability. The rules are currently
// hardcoded.
func (cta *LegacyClientTypeAnalyzer) AgentIsMonitor(rec conversion.InputRecord) bool {
	agentStr := strings.ToLower(rec.GetUserAgent())
	return strings.Index(agentStr, "python-urllib/2.7") > -1 ||
		strings.Index(agentStr, "zabbix-test") > -1
}

func (cta *LegacyClientTypeAnalyzer) AgentIsBot(rec conversion.InputRecord) bool {
	agentStr := strings.ToLower(rec.GetUserAgent())
	// TODO move this to some external file
	return strings.Index(agentStr, "ahrefsbot") > -1 ||
		strings.Index(agentStr, "applebot") > -1 ||
		strings.Index(agentStr, "baiduspider") > -1 ||
		strings.Index(agentStr, "bingbot") > -1 ||
		strings.Index(agentStr, "blexbot") > -1 ||
		strings.Index(agentStr, "dotbot") > -1 ||
		strings.Index(agentStr, "duckduckbot") > -1 ||
		strings.Index(agentStr, "exabot") > -1 ||
		strings.Index(agentStr, "googlebot") > -1 ||
		strings.Index(agentStr, "ia_archiver") > -1 ||
		strings.Index(agentStr, "mail.ru_bot") > -1 ||
		strings.Index(agentStr, "mauibot") > -1 ||
		strings.Index(agentStr, "mediatoolkitbot") > -1 ||
		strings.Index(agentStr, "megaindex.ru") > -1 ||
		strings.Index(agentStr, "mj12bot") > -1 ||
		strings.Index(agentStr, "semanticscholarbot") > -1 ||
		strings.Index(agentStr, "semrushbot") > -1 ||
		strings.Index(agentStr, "seokicks-robot") > -1 ||
		strings.Index(agentStr, "seznambot") > -1 ||
		strings.Index(agentStr, "yacybot") > -1 ||
		strings.Index(agentStr, "yahoo") > -1 && strings.Index(agentStr, "slurp") > -1 ||
		strings.Index(agentStr, "yandexbot") > -1
}

// HasBlacklistedIP in the legacy analyzer cannot use blacklists
// so it always returns false
func (cta *LegacyClientTypeAnalyzer) HasBlacklistedIP(rec conversion.InputRecord) bool {
	return false
}
