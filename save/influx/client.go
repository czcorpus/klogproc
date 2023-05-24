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

package influx

import (
	"fmt"

	"klogproc/conversion"

	"github.com/rs/zerolog/log"

	client "github.com/influxdata/influxdb1-client/v2"
)

const (
	defaultReqTimeoutSecs = 10
)

// ConnectionConf specifies a configuration required to store data
// to an InfluxDB database
type ConnectionConf struct {
	Server          string `json:"server"`
	PushChunkSize   int    `json:"pushChunkSize"`
	Database        string `json:"database"`
	Measurement     string `json:"measurement"`
	RetentionPolicy string `json:"retentionPolicy"`
	ReqTimeoutSecs  int    `json:"reqTimeoutSecs"`
}

// IsConfigured tests whether the configuration is considered
// to be enabled (i.e. no error checking just enabled/disabled)
func (conf *ConnectionConf) IsConfigured() bool {
	return conf.Server != ""
}

// Validate tests whether the configuration is filled in
// correctly. Please note that if the function returns nil
// then IsConfigured() must return 'true'.
func (conf *ConnectionConf) Validate() error {
	var err error
	if conf.Server == "" {
		err = fmt.Errorf("missing 'server' information for InfluxDB")
	}
	if conf.Database == "" {
		err = fmt.Errorf("missing 'database' information for InfluxDB")
	}
	if conf.Measurement == "" {
		err = fmt.Errorf("missing 'measurement' information for InfluxDB")
	}
	if conf.RetentionPolicy == "" {
		err = fmt.Errorf("missing 'retentionPolicy' information for InfluxDB")
	}
	if conf.ReqTimeoutSecs == 0 {
		conf.ReqTimeoutSecs = defaultReqTimeoutSecs
		log.Warn().Msgf("value influxDb.reqTimeoutSecs not specified, using default %d", defaultReqTimeoutSecs)
	}
	return err
}

// ------

func newBatchPoints(database string, retentionPolicy string) (client.BatchPoints, error) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Precision:        "s",
		Database:         database,
		RetentionPolicy:  retentionPolicy,
		WriteConsistency: "one",
	})
	if err != nil {
		return nil, err
	}
	return bp, nil
}

// RecordWriter is a simple wrapper around InfluxDB client allowing
// adding records in a convenient way without need to think
// about batch processing of the records. The price paid here
// is that the client is statefull and Finish() method must
// be always called to finish the current operation.
type RecordWriter struct {
	conn            client.Client
	address         string
	database        string
	retentionPolicy string
	measurement     string
	pushChunkSize   int
	bp              client.BatchPoints
}

// AddRecord adds a record and if internal batch is full then
// it also stores the record to a configured database and
// measurement. Please note that without calling Finish() at
// the end of an operation, stale records may remain.
func (c *RecordWriter) AddRecord(rec conversion.OutputRecord) (bool, error) {
	tags, values := rec.ToInfluxDB()
	point, err := client.NewPoint(c.measurement, tags, values, rec.GetTime())
	if err != nil {
		log.Error().Msgf("Failed to add record to influxdb: %s", err)
	}
	c.bp.AddPoint(point)
	if len(c.bp.Points()) == c.pushChunkSize {
		return true, c.writeCurrBatch()
	}
	return false, nil
}

// Finish ensures that the current operation is fully
// processed and all the data are written to InfluxDB.
func (c *RecordWriter) Finish() error {
	return c.writeCurrBatch()
}

func (c *RecordWriter) writeCurrBatch() error {
	var err error
	err = c.conn.Write(c.bp)
	if err != nil {
		return err
	}
	c.bp, err = newBatchPoints(c.database, c.retentionPolicy)
	if err != nil {
		return err
	}
	return nil
}

// NewRecordWriter is a factory function for RecordWriter
func NewRecordWriter(conf *ConnectionConf) (*RecordWriter, error) {
	conn, err := client.NewHTTPClient(client.HTTPConfig{Addr: conf.Server})
	if err != nil {
		return nil, err
	}

	bp, err := newBatchPoints(conf.Database, conf.RetentionPolicy)
	if err != nil {
		return nil, err
	}

	return &RecordWriter{
		conn:            conn,
		address:         conf.Server,
		database:        conf.Database,
		retentionPolicy: conf.RetentionPolicy,
		measurement:     conf.Measurement,
		bp:              bp,
		pushChunkSize:   conf.PushChunkSize,
	}, nil
}
