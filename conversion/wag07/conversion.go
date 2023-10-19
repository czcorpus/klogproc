// Copyright 2021 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2021 Institute of the Czech National Corpus,
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

package wag07

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"klogproc/conversion"
	"klogproc/conversion/wag06"
	"klogproc/email"
	"klogproc/load"
	"klogproc/logbuffer"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/datetime"
	"github.com/czcorpus/cnc-gokit/maths"
	"github.com/rs/zerolog/log"
)

type sitemsWrapper struct {
	data collections.BinTree[ReqCalcItem]
}

func (w *sitemsWrapper) Get(idx int) maths.FreqInfo {
	return w.data.Get(idx)
}

func (w *sitemsWrapper) Len() int {
	return w.data.Len()
}

type ReqCalcItem struct {
	IP    string
	Count int
	Known bool
}

// Freq is implemented to satisfy cnc-gokit utils
func (rc ReqCalcItem) Freq() int {
	return rc.Count
}

func (rc ReqCalcItem) Compare(other collections.Comparable) int {
	if rc.Count > other.(ReqCalcItem).Count {
		return 1

	} else if rc.Count == other.(ReqCalcItem).Count {
		return 0
	}
	return -1
}

// Transformer converts a source log object into a destination one
type Transformer struct {
	bufferConf    *load.BufferConf
	emailNotifier email.MailNotifier
	realtimeClock bool
}

func (t *Transformer) Transform(logRecord *InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (*wag06.OutputRecord, error) {
	rec := wag06.NewTimedOutputRecord(logRecord.GetTime(), tzShiftMin)
	rec.Type = recType
	rec.Action = logRecord.Action
	rec.IPAddress = logRecord.Request.Origin
	rec.UserAgent = logRecord.Request.HTTPUserAgent
	rec.ReferringDomain = logRecord.Request.Referer
	rec.UserID = strconv.Itoa(logRecord.UserID)
	rec.IsAnonymous = conversion.UserBelongsToList(logRecord.UserID, anonymousUsers)
	rec.IsQuery = logRecord.IsQuery
	rec.IsMobileClient = logRecord.IsMobileClient
	rec.HasPosSpecification = logRecord.HasPosSpecification
	rec.QueryType = logRecord.QueryType
	rec.Lang1 = logRecord.Lang1
	rec.Lang2 = logRecord.Lang2
	rec.Queries = []string{} // no more used?
	rec.ProcTime = -1        // TODO not available; does it have a value
	rec.ID = wag06.CreateID(rec)
	return rec, nil
}

func (t *Transformer) HistoryLookupItems() int {
	return 0
}

func (t *Transformer) isIgnoredIP(ip net.IP) bool {
	return ip == nil || ip.IsLoopback() || ip.IsUnspecified()
}

func (t *Transformer) Preprocess(
	rec conversion.InputRecord, prevRecs logbuffer.AbstractStorage[conversion.InputRecord],
) []conversion.InputRecord {
	tRec, ok := rec.(*InputRecord)
	ans := []conversion.InputRecord{rec}
	if !ok {
		log.Warn().Msg("invalid record passed to wag07.Preprocess()")
		return ans
	}
	if t.bufferConf.BotDetection == nil || (tRec.Action != "search" &&
		tRec.Action != "compare" &&
		tRec.Action != "translate") {
		return ans
	}

	lastCheck := prevRecs.GetTimestamp()
	currTime := rec.GetTime()
	if t.realtimeClock {
		currTime = time.Now()
	}
	ci := time.Duration(t.bufferConf.AnalysisIntervalSecs) * time.Second
	if lastCheck.IsZero() {
		prevRecs.SetTimestamp(currTime)

	} else if rec.GetTime().Sub(lastCheck) > ci {
		defer prevRecs.SetTimestamp(currTime)

		numRec := prevRecs.TotalNumOfRecords()
		sampleSize := prevRecs.AddNumberSample("prevNums", float64(numRec))
		if sampleSize == 1 {
			return ans
		}

		prevNumsRecs := prevRecs.GetNumberSamples("prevNums")
		var meanReqs float64
		for _, v := range prevNumsRecs {
			meanReqs += v
		}
		meanReqs /= float64(len(prevNumsRecs))
		trafficIncrease := float64(numRec) / meanReqs
		var isSuspicTrafficIncrease bool
		if trafficIncrease >= t.bufferConf.BotDetection.TrafficReportingThreshold {
			isSuspicTrafficIncrease = true
			log.Info().
				Str("appType", "wag").
				Float64("prevReqsSampleMean", meanReqs).
				Int("currentReqs", numRec).
				Float64("increase", trafficIncrease).
				Msg("found suspicious increase in traffic - going to report")

			go func() {
				t.emailNotifier.SendFormattedNotification(
					"Klogproc for WaG: suspicious increase in traffic",
					fmt.Sprintf(
						"<p>previous (sampled): <strong>%d</strong>, current: <strong>%d</strong> (increase %01.2f)<br />",
						int(meanReqs), numRec, trafficIncrease),
					fmt.Sprintf("checking interval: <strong>%s</strong><br />", ci.String()),
					fmt.Sprintf("last check: <strong>%v</strong></p>", datetime.FormatDatetime(lastCheck)),
				)
			}()
		}

		counter := make(map[string]ReqCalcItem)
		prevRecs.TotalForEach(func(item conversion.InputRecord) {
			if t.isIgnoredIP(item.GetClientIP()) {
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
			t.bufferConf.BotDetection.IPOutlierMinFreq,
			int(float64(qrt.Q3)+t.bufferConf.BotDetection.IPOutlierCoeff*float64(qrt.IQR())),
		)
		suspiciousRecords := make([]ReqCalcItem, 0, sortedItems.Len()/2)
		sortedItems.ForEach(func(i int, v ReqCalcItem) bool {
			if v.Count > threshold {
				if collections.SliceContains[string](t.bufferConf.BotDetection.BlocklistIP, v.IP) {
					v.Known = true
				}
				suspiciousRecords = append(suspiciousRecords, v)
			}
			return true
		})

		if len(suspiciousRecords) > 0 {

			log.Info().
				Str("appType", "wag").
				Int("threshold", threshold).
				Int("numOutliers", len(suspiciousRecords)).
				Msg("found outlier IP requests - going to report")

			sort.SliceStable(suspiciousRecords, func(i, j int) bool {
				return suspiciousRecords[i].Count > suspiciousRecords[j].Count
			})

			prevRecs.TotalRemoveAnalyzedRecords(lastCheck)
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
				t.emailNotifier.SendFormattedNotification(
					"Klogproc for WaG: suspicious IP addresses detected",
					trafficNote,
					"suspicious records:",
					ipTable.String(),
					fmt.Sprintf("<p>total requesting IPs: <strong>%d</strong><br />", sortedItems.Len()),
					fmt.Sprintf("threshold: %d requests", threshold),
					fmt.Sprintf("last check: <strong>%v</strong><br />", datetime.FormatDatetime(lastCheck)),
					fmt.Sprintf("checking interval: <strong>%s</strong><br /></p>", ci.String()),
				)
			}()
		}
	}

	return ans
}

func NewTransformer(
	bufferConf *load.BufferConf,
	realtimeClock bool,
	emailNotifier email.MailNotifier,
) *Transformer {
	return &Transformer{
		bufferConf:    bufferConf,
		realtimeClock: realtimeClock,
		emailNotifier: emailNotifier,
	}
}
