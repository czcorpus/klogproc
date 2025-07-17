// Copyright 2025 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2025 Institute of the Czech National Corpus,
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

package elastic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type SnapshotMetadata struct {
	TakenBy      string `json:"taken_by"`
	TakenBecause string `json:"taken_because"`
}

type CreateSnapshotArgs struct {
	Indices            string           `json:"indices"`
	IgnoreUnavailable  bool             `json:"ignore_unavailable"`
	IncludeGlobalState bool             `json:"include_global_state"`
	Metadata           SnapshotMetadata `json:"metadata"`
}

type CreateSnapshotResponse struct {
	Snapshot           string           `json:"snapshot"`
	UUID               string           `json:"uuid"`
	Repository         string           `json:"repository"`
	VersionID          any              `json:"version_id"`
	Version            any              `json:"version"`
	Indices            []string         `json:"indices"`
	DataStreams        []string         `json:"data_streams"`
	IncludeGlobalState bool             `json:"include_global_state"`
	FeatureStates      []any            `json:"feature_states"`
	Metadata           SnapshotMetadata `json:"metadata"`
	State              string           `json:"state"`
	StartTime          string           `json:"start_time"`
	StartTimeInMillis  int64            `json:"start_time_in_millis"`
	EndTime            string           `json:"end_time"`
	EndTimeInMillis    int64            `json:"end_time_in_millis"`
	DurationInMillis   int64            `json:"duration_in_millis"`
	Failures           []any            `json:"failures"`
	Shards             struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"shards"`
}

func (c *ESClient) createSnapshot(repository string, name string, query []byte) (*CreateSnapshotResponse, error) {
	path := "/_snapshot/" + repository + "/" + name
	resp, err := c.Do("POST", path, query)
	if err != nil {
		return nil, err
	}
	var snapshotInfo CreateSnapshotResponse
	if err2 := json.Unmarshal(resp, &snapshotInfo); err2 != nil {
		return nil, err2
	}
	return &snapshotInfo, nil
}

// Create snapshot
func (c *ESClient) CreateSnapshot(repository string, name string, reason string) (*CreateSnapshotResponse, error) {
	snapshotArgs := CreateSnapshotArgs{
		Indices:            c.index,
		IgnoreUnavailable:  false,
		IncludeGlobalState: false,
		Metadata: SnapshotMetadata{
			TakenBy:      "Klogproc",
			TakenBecause: reason,
		},
	}
	query, err := json.Marshal(snapshotArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot args: %w", err)
	}

	log.Debug().Any("snapshotArgs", snapshotArgs).Msg("Creating snapshot")
	if name == "" {
		name = fmt.Sprintf("snapshot_%d", time.Now().Unix())
	}
	return c.createSnapshot(repository, name, query)
}

type ListSnapshotsResponse struct {
	Snapshots []struct {
		Snapshot          string   `json:"snapshot"`
		UUID              string   `json:"uuid"`
		VersionID         int64    `json:"version_id"`
		Version           string   `json:"version"`
		Indices           []string `json:"indices"`
		State             string   `json:"state"`
		StartTime         string   `json:"start_time"`
		StartTimeInMillis int64    `json:"start_time_in_millis"`
		EndTime           string   `json:"end_time"`
		EndTimeInMillis   int64    `json:"end_time_in_millis"`
		DurationInMillis  int64    `json:"duration_in_millis"`
		Failures          []any    `json:"failures"`
		Shards            struct {
			Total      int `json:"total"`
			Successful int `json:"successful"`
			Failed     int `json:"failed"`
		} `json:"shards"`
	} `json:"snapshots"`
}

func (c *ESClient) ListSnapshots(repository string) (*ListSnapshotsResponse, error) {
	path := "/_snapshot/" + repository + "/_all"
	resp, err := c.Do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("ListSnapshots response: %s", string(resp))
	var snapshotList ListSnapshotsResponse
	if err2 := json.Unmarshal(resp, &snapshotList); err2 != nil {
		return nil, err2
	}
	return &snapshotList, nil
}

func (c *ESClient) RemoveSnapshot(repository string, name string) error {
	path := "/_snapshot/" + repository + "/" + name
	_, err := c.Do("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %w", name, err)
	}
	return nil
}

type RestoreSnapshotArgs struct {
	Indices            string `json:"indices"`
	IgnoreUnavailable  bool   `json:"ignore_unavailable"`
	IncludeGlobalState bool   `json:"include_global_state"`
}

func (c *ESClient) RestoreSnapshot(repository string, name string) error {
	path := "/_snapshot/" + repository + "/" + name + "/_restore"
	snapshotArgs := RestoreSnapshotArgs{
		Indices:            c.index,
		IgnoreUnavailable:  false,
		IncludeGlobalState: false,
	}
	query, err := json.Marshal(snapshotArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot args: %w", err)
	}
	_, err = c.Do("POST", path, query)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot %s: %w", name, err)
	}
	return nil
}
