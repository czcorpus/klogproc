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
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
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

type repoSettings struct {
	Compress bool   `json:"compress"`
	Location string `json:"location"`
}

type repoInfo struct {
	Type     string       `json:"type"`
	Settings repoSettings `json:"settings"`
}

type repoListResponse map[string]repoInfo

type acceptedResp struct {
	Accepted bool `json:"accepted"`
}

func (c *ESClient) createSnapshot(repository, name string, args CreateSnapshotArgs) (string, error) {
	url := fmt.Sprintf("/_snapshot/%s/%s", repository, name)
	query, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}
	log.Debug().Any("snapshotArgs", args).Msg("Creating snapshot")
	resp, err := c.DoRequest("PUT", url, query)
	if err != nil {
		return "", err
	}
	var respObj acceptedResp
	if err2 := json.Unmarshal(resp, &respObj); err2 != nil {
		return "", err2
	}
	if respObj.Accepted {
		return name, nil
	}
	return "", fmt.Errorf("snapshot %s not accepted", name)
}

func (c *ESClient) testRepoExists(repository, rootFSPath string) (bool, error) {
	url := fmt.Sprintf("/_snapshot/%s", repository)
	_, err := c.DoRequest("GET", url, []byte{})
	if err != nil {
		var tErr *ESClientError
		if errors.As(err, &tErr) {
			if tErr.ESError.Status == http.StatusNotFound {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to test repository existence: %w", err)
	}
	return true, nil
}

func (c *ESClient) createSnapshotRepo(repository, rootFSPath string) error {
	exists, err := c.testRepoExists(repository, rootFSPath)
	if err != nil {
		return fmt.Errorf("failed to creeate snapshot repository: %w", err)
	}
	if exists {
		log.Info().
			Str("appType", repository).
			Msg("repository for snapshots already exists, using that one")
		return nil
	}

	url := fmt.Sprintf("/_snapshot/%s", repository)
	body, err := json.Marshal(repoInfo{
		Type: "fs",
		Settings: repoSettings{
			Compress: true,
			Location: filepath.Join(rootFSPath, repository),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshot repository: %w", err)
	}
	if _, err := c.DoRequest("PUT", url, body); err != nil {
		return fmt.Errorf("failed to create snapshot repository: %w", err)
	}
	return nil
}

// Create snapshot
func (c *ESClient) CreateSnapshot(conf SnapshotConf, appType, name, reason string) (string, error) {
	if err := c.createSnapshotRepo(appType, conf.RootFSPath); err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}
	if name == "" {
		name = fmt.Sprintf("snapshot_%s", time.Now().Format("2006-01-02-15-04-05"))
	}
	if reason == "" {
		reason = "Unspecified"
	}
	return c.createSnapshot(
		appType,
		name,
		CreateSnapshotArgs{
			Indices:            c.index,
			IgnoreUnavailable:  false,
			IncludeGlobalState: false,
			Metadata: SnapshotMetadata{
				TakenBy:      "Klogproc",
				TakenBecause: reason,
			},
		},
	)
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

func (c *ESClient) ListSnapshots(conf SnapshotConf, appType string) (*ListSnapshotsResponse, error) {
	path := "/_snapshot/" + appType + "/_all"
	resp, err := c.DoRequest("GET", path, nil)
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

func (c *ESClient) RemoveSnapshot(conf SnapshotConf, appType, name string) error {
	path := "/_snapshot/" + appType + "/" + name
	_, err := c.DoRequest("DELETE", path, nil)
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

func (c *ESClient) RestoreSnapshot(conf SnapshotConf, appType, name string) error {
	path := "/_snapshot/" + appType + "/" + name + "/_restore"
	snapshotArgs := RestoreSnapshotArgs{
		Indices:            c.index,
		IgnoreUnavailable:  false,
		IncludeGlobalState: false,
	}
	query, err := json.Marshal(snapshotArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot args: %w", err)
	}
	_, err = c.DoRequest("POST", path, query)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot %s: %w", name, err)
	}
	return nil
}
