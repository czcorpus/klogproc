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
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/datetime"
	"github.com/czcorpus/cnc-gokit/maths"
	"github.com/rs/zerolog/log"
)

const (
	minPrevNumRequestsSampleSize = 10
	bufferCleanupProbability     = 0.1
	bufferCleanupMaxAge          = time.Hour * 6
	suspiciousRecordsThreshold   = 0.6
	suspiciousRecordsMinRequests = 20
	fullBufferMaxAge             = time.Hour * 5
)

type SuspiciousReqCounter struct {
	NumAny    int
	NumSuspic int
	LastUpd   time.Time
}

func (src SuspiciousReqCounter) SuspicRatio() float64 {
	return float64(src.NumSuspic) / float64(src.NumAny)
}

type IPReport struct {
	IP   string `json:"ip"`
	Freq int    `json:"freq"`
}

// BotAnalysisState contains values helpful to determine
// suspicious traffic in a log.
type BotAnalysisState struct {
	PrevNums          *logbuffer.SampleWithReplac[int]                         `json:"prevNums"`
	LastCheck         time.Time                                                `json:"timestamp"`
	TotalProcessed    int                                                      `json:"totalProcessed"`
	FullBufferIPProps *collections.ConcurrentMap[string, SuspiciousReqCounter] `json:"fullBufferIPProps"`
}

func (state *BotAnalysisState) ToJSON() ([]byte, error) {
	return json.Marshal(state)
}

func (state *BotAnalysisState) AfterLoadNormalize(conf *load.BufferConf, dt time.Time) {
	if state.LastCheck.IsZero() && state.PrevNums.Len() > 0 {
		state.LastCheck = dt

	} else if state.PrevNums.Cap == 0 {
		state.PrevNums = logbuffer.NewSampleWithReplac[int](state.PrevNums.Cap)
	}
	if state.PrevNums.Cap != conf.BotDetection.PrevNumReqsSampleSize {
		state.PrevNums.Resize(conf.BotDetection.PrevNumReqsSampleSize)
	}
	if state.FullBufferIPProps == nil {
		state.FullBufferIPProps = collections.NewConcurrentMap[string, SuspiciousReqCounter]()
	}
}

func (state *BotAnalysisState) Report() map[string]any {
	ans := make(map[string]any)
	var pnm float64
	for _, v := range state.PrevNums.GetAll() {
		pnm += float64(v)
	}
	ans["prevNumsMean"] = pnm / float64(state.PrevNums.Len())
	ans["prevNumsLen"] = state.PrevNums.Len()
	ans["lastCheck"] = state.LastCheck
	ans["totalProcessed"] = state.TotalProcessed
	return ans
}

// BotAnalyzer is used in the "preprocess" phase of
// the servicelog.process. Using log records buffer,
// it searches for suspicious traffic and IP addresses.
//
// Basically, it looks for
// 1) high increase in traffic (a respective ratio is defined in config)
// 2) outlier IPs with too high ratio on the recent traffic
//
// Both are reported via e-mail.
type BotAnalyzer[T AnalyzableRecord] struct {
	appType       string
	conf          *load.BufferConf
	realtimeClock bool
	emailNotifier email.MailNotifier
}

func (analyzer *BotAnalyzer[T]) isIgnoredIP(ip net.IP) bool {
	return ip == nil || ip.IsLoopback() || ip.IsUnspecified()
}

func (analyzer *BotAnalyzer[T]) Preprocess(
	rec servicelog.InputRecord,
	prevRecs logbuffer.AbstractStorage[servicelog.InputRecord, logbuffer.SerializableState],
) []servicelog.InputRecord {

	currTime := rec.GetTime()
	if analyzer.realtimeClock {
		currTime = time.Now()
	}

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
	state := prevRecs.GetStateData(currTime)
	tState, ok := state.(*BotAnalysisState)
	if !ok {
		log.Error().Str("appType", analyzer.appType).Msg("invalid analysis state type for")
		return ans
	}
	defer prevRecs.SetStateData(tState)
	tState.TotalProcessed++

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
			err := analyzer.emailNotifier.SendNotification(
				fmt.Sprintf(
					"Klogproc for %s: suspicious increase in traffic", analyzer.appType),
				map[string]any{},
				fmt.Sprintf(
					"previous (sampled): **%d**, current: **%d** (increase %01.2f)  ",
					int(meanReqs), numRec, trafficIncrease),
				fmt.Sprintf("checking interval: **%s**  ", checkInterval.String()),
				fmt.Sprintf("last check: **%v**", datetime.FormatDatetime(tState.LastCheck)),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to send notification")
			}
		}()
	}

	lastPeriodCounter := make(map[string]*ReqCalcItem)
	var avgRequests float64
	prevRecs.TotalForEach(func(item servicelog.InputRecord) {
		if analyzer.isIgnoredIP(item.GetClientIP()) || !item.GetTime().After(tState.LastCheck) {
			return
		}
		curr, ok := lastPeriodCounter[item.GetClientIP().String()]
		if !ok {
			curr = &ReqCalcItem{
				IP: item.GetClientIP().String(),
			}
		}
		curr.Count++
		avgRequests += 1
		suspicCnt := tState.FullBufferIPProps.Get(item.GetClientIP().String())
		suspicCnt.NumAny++
		suspicCnt.LastUpd = currTime
		if item.IsSuspicious() {
			suspicCnt.NumSuspic++
		}
		tState.FullBufferIPProps.Set(item.GetClientIP().String(), suspicCnt)
		lastPeriodCounter[item.GetClientIP().String()] = curr
	})
	avgRequests /= float64(len(lastPeriodCounter))
	sortedItems := collections.BinTree[*ReqCalcItem]{}
	suspicRequestsIP := collections.Set[string]{}
	for _, v := range lastPeriodCounter {
		sortedItems.Add(v)
		fullBufferInfo := tState.FullBufferIPProps.Get(v.IP)
		if fullBufferInfo.SuspicRatio() >= suspiciousRecordsThreshold &&
			v.Count >= int(avgRequests) {
			suspicRequestsIP.Add(v.IP)
		}
	}

	var numCleaned int
	tState.FullBufferIPProps.Filter(func(k string, v SuspiciousReqCounter) bool {
		if v.LastUpd.Before(currTime.Add(-fullBufferMaxAge)) {
			numCleaned++
			return false
		}
		return true
	})
	if numCleaned > 0 {
		log.Info().Int("numCleaned", numCleaned).Msg("cleaned old records in full buffer IP props table")
	}

	if suspicRequestsIP.Size() > 0 {
		go func() {
			err := analyzer.emailNotifier.SendNotification(
				fmt.Sprintf("Klogproc for %s: suspicious IP addresses detected", analyzer.appType),
				map[string]any{},
				"records with high ratio of suspicious requests:",
				strings.Join(suspicRequestsIP.ToSlice(), "  \n"),
				fmt.Sprintf("total requesting IPs: **%d**  ", sortedItems.Len()),
				fmt.Sprintf("last check: **%v**  ", datetime.FormatDatetime(tState.LastCheck)),
				fmt.Sprintf("checking interval: **%s**  ", checkInterval.String()),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to send notification")
			}
		}()
	}

	qrt, err := maths.GetQuartiles[maths.FreqInfo](&sitemsWrapper{sortedItems})
	if err == maths.ErrTooSmallDataset {
		return ans
	}
	threshold := maths.Max(
		analyzer.conf.BotDetection.IPOutlierMinFreq,
		int(float64(qrt.Q3)+analyzer.conf.BotDetection.IPOutlierCoeff*float64(qrt.IQR())),
	)
	outlierRecords := make([]*ReqCalcItem, 0, sortedItems.Len()/2)
	sortedItems.ForEach(func(i int, v *ReqCalcItem) bool {
		if v.Count > threshold {
			if collections.SliceContains[string](analyzer.conf.BotDetection.BlocklistIP, v.IP) {
				v.Known = true
			}
			outlierRecords = append(outlierRecords, v)
		}
		return true
	})

	if len(outlierRecords) > 0 {

		log.Info().
			Str("appType", analyzer.appType).
			Int("threshold", threshold).
			Int("numOutliers", len(outlierRecords)).
			Msg("found outlier IP requests - going to report")

		sort.SliceStable(outlierRecords, func(i, j int) bool {
			return outlierRecords[i].Count > outlierRecords[j].Count
		})

		ipReportMetadata := make([]IPReport, len(outlierRecords))
		var ipListing strings.Builder
		for i, susp := range outlierRecords {
			ipListing.WriteString(fmt.Sprintf("%s (%dx)\\n", susp.IP, susp.Count))
			ipReportMetadata[i] = IPReport{IP: susp.IP, Freq: susp.Count}
		}

		var trafficNote string
		if isSuspicTrafficIncrease {
			trafficNote = fmt.Sprintf(
				"**This report is supported with suspicious increase of traffic (%01.2f).**",
				trafficIncrease,
			)
		}

		go func() {
			err := analyzer.emailNotifier.SendNotification(
				fmt.Sprintf("Klogproc for %s: suspicious IP addresses detected", analyzer.appType),
				map[string]any{"ipList": ipReportMetadata},
				trafficNote,
				"suspicious records:",
				ipListing.String(),
				fmt.Sprintf("total requesting IPs: **%d**  ", sortedItems.Len()),
				fmt.Sprintf("threshold: %d requests  ", threshold),
				fmt.Sprintf("last check: **%v**  ", datetime.FormatDatetime(tState.LastCheck)),
				fmt.Sprintf("checking interval: **%s**  ", checkInterval.String()),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to send notification")
			}
		}()
	}

	if rand.Float64() < bufferCleanupProbability {
		limitDt := time.Now().Add(-bufferCleanupMaxAge)
		numRm := prevRecs.ClearOldRecords(limitDt)
		log.Info().
			Int("numRemoved", numRm).
			Int("lenghtAfter", prevRecs.TotalNumOfRecordsSince(limitDt)).
			Msg("performed buffer records cleanup")
	}

	return ans
}

func NewBotAnalyzer[T AnalyzableRecord](
	appType string,
	conf *load.BufferConf,
	realtimeClock bool,
	emailNotifier email.MailNotifier,
) *BotAnalyzer[T] {
	return &BotAnalyzer[T]{
		appType:       appType,
		conf:          conf,
		realtimeClock: realtimeClock,
		emailNotifier: emailNotifier,
	}
}
