// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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

package analysis

import (
	"encoding/json"
	"fmt"
	"klogproc/email"
	"klogproc/load"
	"klogproc/logbuffer"
	"klogproc/servicelog"
	"net"
	"sort"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/datetime"
	"github.com/czcorpus/cnc-gokit/maths"
	"github.com/rs/zerolog/log"
)

const (
	minPrevNumRequestsSampleSize = 10
)

type AnalysisState struct {
	PrevNums  logbuffer.SampleWithReplac[int] `json:"prevNums"`
	LastCheck time.Time                       `json:"timestamp"`
}

// while looking superfluous, we need this to fullfill a required interface
func (state *AnalysisState) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrevNums  logbuffer.SampleWithReplac[int] `json:"prevNums"`
		LastCheck time.Time                       `json:"timestamp"`
	}{
		PrevNums:  state.PrevNums,
		LastCheck: state.LastCheck,
	})
}

// Analyzer is used in the "preprocess" phase of
// the servicelog.process. Using log records buffer,
// it searches for suspicious traffic and IP addresses.
//
// Basically, it looks for
// 1) high increase in traffic (a respective ratio is defined in config)
// 2) outlier IPs with too high ratio on the recent traffic
//
// Both are reported via e-mail.
type Analyzer[T AnalyzableRecord] struct {
	appType       string
	conf          *load.BufferConf
	realtimeClock bool
	emailNotifier email.MailNotifier
}

func (analyzer *Analyzer[T]) isIgnoredIP(ip net.IP) bool {
	return ip == nil || ip.IsLoopback() || ip.IsUnspecified()
}

func (analyzer *Analyzer[T]) Preprocess(
	rec servicelog.InputRecord, prevRecs logbuffer.AbstractStorage[servicelog.InputRecord, logbuffer.SerializableState],
) []servicelog.InputRecord {
	tRec, ok := rec.(T)
	ans := []servicelog.InputRecord{rec}
	if !ok {
		log.Warn().
			Str("appType", analyzer.appType).
			Msg("invalid record type passed to Analyzer")
		return ans
	}
	if analyzer.conf.BotDetection == nil || !tRec.ShouldBeAnalyzed() {
		return ans
	}
	state := prevRecs.GetStateData()
	tState, ok := state.(*AnalysisState)
	if !ok {
		log.Error().Str("appType", analyzer.appType).Msg("invalid analysis state type for")
		return ans
	}
	defer prevRecs.SetStateData(tState)

	currTime := rec.GetTime()
	if analyzer.realtimeClock {
		currTime = time.Now()
	}

	if tState.LastCheck.IsZero() {
		tState.LastCheck = currTime
		return ans
	}

	checkInterval := time.Duration(analyzer.conf.AnalysisIntervalSecs) * time.Second
	if rec.GetTime().Sub(tState.LastCheck) < checkInterval {
		return ans
	}

	defer func() { tState.LastCheck = currTime }()
	numRec := prevRecs.TotalNumOfRecordsSince(tState.LastCheck)
	sampleSize := tState.PrevNums.Add(numRec)
	if sampleSize < minPrevNumRequestsSampleSize {
		log.Debug().
			Int("numRec", numRec).
			Int("sampleSize", sampleSize).
			Int("sampleLimit", minPrevNumRequestsSampleSize).
			Msg("previous requests sample not ready yet")
		return ans
	}

	var meanReqs float64
	for _, v := range tState.PrevNums.Data {
		meanReqs += float64(v)
	}
	meanReqs /= float64(len(tState.PrevNums.Data))
	trafficIncrease := float64(numRec) / meanReqs
	log.Debug().
		Time("lastCheck", tState.LastCheck).
		Int("prevNumRecsSampleSize", len(tState.PrevNums.Data)).
		Float64("meanPrevReqs", meanReqs).
		Int("numRec", numRec).
		Float64("trafficIncrease", trafficIncrease).
		Str("appType", analyzer.appType).
		Msg("Checking for suspicious activity")
	var isSuspicTrafficIncrease bool
	if trafficIncrease >= analyzer.conf.BotDetection.TrafficReportingThreshold {
		isSuspicTrafficIncrease = true
		log.Info().
			Str("appType", analyzer.appType).
			Float64("prevReqsSampleMean", meanReqs).
			Int("currentReqs", numRec).
			Float64("increase", trafficIncrease).
			Msg("found suspicious increase in traffic - going to report")

		go func() {
			analyzer.emailNotifier.SendFormattedNotification(
				fmt.Sprintf(
					"Klogproc for %s: suspicious increase in traffic", analyzer.appType),
				fmt.Sprintf(
					"<p>previous (sampled): <strong>%d</strong>, current: <strong>%d</strong> (increase %01.2f)<br />",
					int(meanReqs), numRec, trafficIncrease),
				fmt.Sprintf("checking interval: <strong>%s</strong><br />", checkInterval.String()),
				fmt.Sprintf("last check: <strong>%v</strong></p>", datetime.FormatDatetime(tState.LastCheck)),
			)
		}()
	}

	counter := make(map[string]ReqCalcItem)
	prevRecs.TotalForEach(func(item servicelog.InputRecord) {
		if analyzer.isIgnoredIP(item.GetClientIP()) || !item.GetTime().After(tState.LastCheck) {
			return
		}
		curr, ok := counter[item.GetClientIP().String()]
		if !ok {
			curr = ReqCalcItem{
				IP: item.GetClientIP().String(),
			}
		}
		curr.Count++
		counter[item.GetClientIP().String()] = curr
	})
	sortedItems := collections.BinTree[ReqCalcItem]{}
	for _, v := range counter {
		sortedItems.Add(v)
	}

	qrt, err := maths.GetQuartiles[maths.FreqInfo](&sitemsWrapper{sortedItems})
	if err == maths.ErrTooSmallDataset {
		return ans
	}
	threshold := maths.Max(
		analyzer.conf.BotDetection.IPOutlierMinFreq,
		int(float64(qrt.Q3)+analyzer.conf.BotDetection.IPOutlierCoeff*float64(qrt.IQR())),
	)
	suspiciousRecords := make([]ReqCalcItem, 0, sortedItems.Len()/2)
	sortedItems.ForEach(func(i int, v ReqCalcItem) bool {
		if v.Count > threshold {
			if collections.SliceContains[string](analyzer.conf.BotDetection.BlocklistIP, v.IP) {
				v.Known = true
			}
			suspiciousRecords = append(suspiciousRecords, v)
		}
		return true
	})

	if len(suspiciousRecords) > 0 {

		log.Info().
			Str("appType", analyzer.appType).
			Int("threshold", threshold).
			Int("numOutliers", len(suspiciousRecords)).
			Msg("found outlier IP requests - going to report")

		sort.SliceStable(suspiciousRecords, func(i, j int) bool {
			return suspiciousRecords[i].Count > suspiciousRecords[j].Count
		})

		numCellStyle := "text-align: right; border: 1px solid #777777; padding: 0.2em 0.4em"
		cellStyle := "border: 1px solid #777777; padding: 0.2em 0.4em"
		thStyle := "text-align: left; padding: 0.2em 0.4em"

		ipTable := new(email.Table)
		tbody := ipTable.
			Init("border-collapse: collapse").
			AddBody().
			AddTR().
			AddTH("IP", thStyle).
			AddTH("req. count", thStyle).
			AddTH("known", thStyle).Close()
		for _, susp := range suspiciousRecords {
			tbody.AddTR().
				AddTD(susp.IP, cellStyle).
				AddTD(fmt.Sprintf("%d", susp.Count), numCellStyle).
				AddTD(fmt.Sprintf("%t", susp.Known), numCellStyle).
				Close()
		}

		var trafficNote string
		if isSuspicTrafficIncrease {
			trafficNote = fmt.Sprintf(
				"<strong>This report is supported with suspicious increase of traffic (%01.2f).</strong>",
				trafficIncrease,
			)
		}

		go func() {
			analyzer.emailNotifier.SendFormattedNotification(
				fmt.Sprintf("Klogproc for %s: suspicious IP addresses detected", analyzer.appType),
				trafficNote,
				"suspicious records:",
				ipTable.String(),
				fmt.Sprintf("<p>total requesting IPs: <strong>%d</strong><br />", sortedItems.Len()),
				fmt.Sprintf("threshold: %d requests", threshold),
				fmt.Sprintf("last check: <strong>%v</strong><br />", datetime.FormatDatetime(tState.LastCheck)),
				fmt.Sprintf("checking interval: <strong>%s</strong><br /></p>", checkInterval.String()),
			)
		}()
	}
	return ans
}

func NewAnalyzer[T AnalyzableRecord](
	appType string,
	conf *load.BufferConf,
	realtimeClock bool,
	emailNotifier email.MailNotifier,
) *Analyzer[T] {
	return &Analyzer[T]{
		appType:       appType,
		conf:          conf,
		realtimeClock: realtimeClock,
		emailNotifier: emailNotifier,
	}
}
