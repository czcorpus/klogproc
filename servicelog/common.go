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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"klogproc/logbuffer"
	"net"
	"strconv"
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (

	// AppTypeAkalex defines a universal storage identifier for Calc
	AppTypeAkalex = "akalex"

	// AppTypeAPIGuard represents a universal storage identifier for APIGuard
	AppTypeAPIGuard = "apiguard"

	// AppTypeCalc defines a universal storage identifier for Calc
	AppTypeCalc = "calc"

	// AppTypeGramatikat defines a universal storage identifier for Gramatikat
	AppTypeGramatikat = "gramatikat"

	// AppTypeKontext defines a universal storage identifier for KonText
	AppTypeKontext = "kontext"

	// AppTypeKontextAPI defines a universal storage identifier for KonText API instance
	AppTypeKontextAPI = "kontext-api"

	// AppTypeKorpusDB defines a universal storage identifier for KorpusDB
	AppTypeKorpusDB = "korpus-db"

	// AppTypeKwords defines a universal storage identifier for Kwords
	AppTypeKwords = "kwords"

	// AppTypeLists defines a universal storage identifier for Lists
	AppTypeLists = "lists"

	// AppTypeMapka defines a universal storage identifier for Mapka
	AppTypeMapka = "mapka"

	// AppTypeMorfio defines a universal storage identifier for Morfio
	AppTypeMorfio = "morfio"

	// AppTypeQuitaUp defines a universal storage identifier for QuitaUp
	AppTypeQuitaUp = "quita-up"

	// AppTypeSke defines a universal storage identifier for Treq
	AppTypeSke = "ske"

	// AppTypeSyd defines a universal storage identifier for SyD
	AppTypeSyd = "syd"

	// AppTypeTreq defines a universal storage identifier for Treq
	AppTypeTreq = "treq"

	// AppTypeWag defines a universal storage identifier for Word at a Glance
	AppTypeWag = "wag"

	// AppTypeWsserver defines a universal storage identifier for Word-Sim-Server
	AppTypeWsserver = "wsserver"

	// AppTypeMasm defines a universal storage identifier for Masm
	AppTypeMasm = "masm"

	// AppTypeMquery defines a universal storage identifier for Mquery
	AppTypeMquery = "mquery"

	// AppTypeMquerySRU defines a universal storage identifier for Mquery-SRU
	AppTypeMquerySRU = "mquery-sru"
)

type ServiceLogBuffer logbuffer.AbstractRecentRecords[InputRecord, logbuffer.SerializableState]

// LineParsingError informs that we failed to parse a line as
// an standard log record. In general, this may or may not mean
// that the line actually contains a broken (= non parseable) string.
// So it is ok to perform additional processing based on the format of
// the logged data in such cases.
// E.g. KonText produces JSON lines for normal operations but in
// case an error occurs, it just dumps a standard Python error
// message along with stack trace (multi-line).
type LineParsingError struct {
	LineNumber int64
	Message    string
}

func (m LineParsingError) Error() string {
	return fmt.Sprintf("%s: LineParsingError at line %d", m.Message, m.LineNumber)
}

// NewLineParsingError is a constructor for LineParsingError
func NewLineParsingError(lineNumber int64, message string) LineParsingError {
	return LineParsingError{LineNumber: lineNumber, Message: message}
}

type StreamedLineParsingError struct {
	RecordPrefix string
	Message      string
}

func (m StreamedLineParsingError) Error() string {
	return fmt.Sprintf("%s: StreamedLineParsingError at \"%s...\"", m.Message, m.RecordPrefix)
}

// NewStreamedLineParsingError is a constructor for StreamedLineParsingError
func NewStreamedLineParsingError(line string, message string) StreamedLineParsingError {
	var sample string
	if len(line) > 35 {
		sample = line[:35]

	} else {
		sample = line
	}
	return StreamedLineParsingError{RecordPrefix: sample, Message: message}
}

func GenerateRandomClusteringID() string {
	id := uuid.New()
	sum := sha1.New()
	_, err := sum.Write([]byte(id.String()))
	if err != nil {
		log.Error().Err(err).Msg("problem generating hash")
	}
	return hex.EncodeToString(sum.Sum(nil))
}

// InputRecord describes a common behavior for objects extracted
// from an application log of any UCNK app.
type InputRecord interface {
	GetTime() time.Time
	GetClientIP() net.IP
	GetUserAgent() string

	// ClusteringClientID should provide the best available identifier of a user
	// usable for requests clustering.
	// The priority is as follows:
	// 1) user ID
	// 2) session ID
	// 3) IP address
	// Please note that the values do not have to be directly the ones listed above.
	// It is perfectly OK to hash the original values.
	ClusteringClientID() string
	ClusterSize() int
	SetCluster(size int)
	IsProcessable() bool

	// IsSuspicious is just a hint for further analysis. It should not be a direct
	// signal to ban the requesting side
	IsSuspicious() bool
}

// GeoDataRecord represents a full client geographical
// position information as provided by GeoIP database
type GeoDataRecord struct {
	ContinentCode string     `json:"continent_code"`
	CountryCode2  string     `json:"country_code2"`
	CountryCode3  string     `json:"country_code3"`
	CountryName   string     `json:"country_name"`
	IP            string     `json:"ip"`
	Latitude      float32    `json:"latitude"`
	Longitude     float32    `json:"longitude"`
	Location      [2]float32 `json:"location"`
	Timezone      string     `json:"timezone"`
}

// OutputRecord describes a common behavior for records ready to
// be stored to the storage with a defined type. Implementation
// details are up to concrete implementations but these functions are
// required by the 'processing template'.
type OutputRecord interface {
	SetLocation(countryName string, latitude float32, longitude float32, timezone string)

	// ToJSON creates an object suitable for storing to ElasticSearch, CouchDB and other
	// document-oriented databases
	ToJSON() ([]byte, error)

	// ToInfluxDB creates two maps: 1) tags, 2) values as defined
	// by InfluxDB architecture. These can be directly saved via
	// a respective InfluxDB client.
	ToInfluxDB() (tags map[string]string, values map[string]interface{})

	// Create an idempotent unique identifier of the record.
	// This can be typically acomplished by hashing the original
	// log record.
	GetID() string

	// Return app type as defined by an external convention
	// (e.g. for UCNK: kontext, syd, morfio, treq,...)
	GetType() string

	// Get time of the log record
	GetTime() time.Time
}

type LogRange struct {
	Inode     int64 `json:"inode"`
	SeekStart int64 `json:"seekStart"`
	SeekEnd   int64 `json:"seekEnd"`
	Written   bool  `json:"written"`
}

func (p LogRange) String() string {
	return fmt.Sprintf("LogRange{Inode: %d, Seek: %d-%d, Written: %t}",
		p.Inode, p.SeekStart, p.SeekEnd, p.Written)
}

type BoundOutputRecord struct {
	Rec      OutputRecord
	FilePos  LogRange
	FilePath string
}

func (r *BoundOutputRecord) ToJSON() ([]byte, error) {
	return r.Rec.ToJSON()
}

func (r *BoundOutputRecord) GetTime() time.Time {
	return r.Rec.GetTime()
}

func (r *BoundOutputRecord) GetID() string {
	return r.Rec.GetID()
}

func (r *BoundOutputRecord) GetType() string {
	return r.Rec.GetType()
}

// LogItemTransformer defines a general object able to transform
// an input log record to an output one.
type LogItemTransformer interface {

	// HistoryLookupItems provides information about whether
	// the transformer needs to look up to previously seen records.
	// This can be used e.g. when clustering records.
	//
	// How the values are understood:
	// x = 0: no buffer, i.e. the transformer will always gets only the current log line
	// x > 0: up to `x` previous records will be available along with the current one
	// x < 0: illegal value
	HistoryLookupItems() int

	Preprocess(rec InputRecord, prevRecs ServiceLogBuffer) []InputRecord

	Transform(logRec InputRecord, recType string, tzShiftMin int, anonymousUsers []int) (OutputRecord, error)
}

// AppErrorRegister describes a type which reacts to logged errors
// (i.e. errors reported by respective applications we watch - not log
// processing errors).
type AppErrorRegister interface {

	// OnError is called whenever a respective parser encounters a reported error
	OnError(message string)

	// Evaluate asks for the current status evaluation and reaction
	// (e.g. an alarm may notify users)
	Evaluate()

	// Reset() should clear internal data (e.g. counters) so it can
	// start again.
	Reset()
}

// UserBelongsToList tests whether a provided user can be
// found in a provided array of users.
func UserBelongsToList[T int | string](userID T, anonymousUsers []int) bool {
	var intUserID int
	var err error
	switch t := any(userID).(type) {
	case int:
		intUserID = t
	case string:
		intUserID, err = strconv.Atoi(t)
		if err != nil {
			return false
		}
	default:
		return false
	}
	for _, v := range anonymousUsers {
		if v == intUserID {
			return true
		}
	}
	return false
}

// Preprocessor is any type able to handle the "preprocess" phase
// of a record processing. It has the following special privileges:
// 1) take an input record (arg `rec`) and return one or more (possibly different) records.
// 2) access to a (config defined) number of previous records
//
// In case it is just a reporting-oriented implementation, the returned
// slice will typically contain just a single item - the original processed
// record. In some cases (see mapka3), where a clustering is required, it
// may take the current record and some of the previous records and define
// whole new items to return.
type Preprocessor interface {
	Preprocess(
		rec InputRecord,
		prevRecs logbuffer.AbstractRecentRecords[InputRecord, logbuffer.SerializableState],
	) []InputRecord
}

// ExcludeIPList represents a list of IP addresses
// which should not be included in log processing
// and archiving. These are typically requests from
// watchdog services.
type ExcludeIPList []string

// Excludes tests an input record whether it should
// be excluded based in its IP address.
func (elist ExcludeIPList) Excludes(rec InputRecord) bool {
	excludes := collections.SliceContains(elist, rec.GetClientIP().String())
	if excludes {
		log.Debug().Str("ip", rec.GetClientIP().String()).Msg("excluded IP")
	}
	return excludes
}

// ErrorRecord specifies a thrown error along with
// optional anchor for easier search within text file
// log
type ErrorRecord struct {
	Name   string `json:"name"`
	Anchor string `json:"anchor"`
}
