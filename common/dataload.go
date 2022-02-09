// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
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

package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func loadHTTPResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("configuration resource loading error: %s (url: %s)", resp.Status, url)
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LoadSupportedResource loads raw byte data for Klogproc configuration.
// Allowed formats are:
// 1) http://..., https://...
// 2) file:/localhost/..., file:///...
// 3) /abs/fs/path, rel/fs/path
func LoadSupportedResource(uri string) ([]byte, error) {
	if uri == "" {
		return nil, fmt.Errorf("no resource (http, file) specified")
	}
	var rawData []byte
	var err error
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		rawData, err = loadHTTPResource(uri)

	} else if strings.HasPrefix(uri, "file:/localhost/") {
		rawData, err = ioutil.ReadFile(uri[len("file:/localhost/")-1:])

	} else if strings.HasPrefix(uri, "file:///") {
		rawData, err = ioutil.ReadFile(uri[len("file:///")-1:])

	} else { // we assume a common fs path
		rawData, err = ioutil.ReadFile(uri)
	}
	if err != nil {
		return nil, err
	}
	return rawData, nil
}
