// © 2023 Snyk Limited All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"

	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/models"
)

// cloudScanDetector is a simple Detector to load cloud scan files.  These
// are usually generated using `snyk iac scaffold spec`.
type cloudScanDetector struct {
}

type cloudScan struct {
	path  string
	state models.State
}

func (cloudScanDetector) DetectFile(file *input.File, opts input.DetectOptions) (input.IACConfiguration, error) {
	if file.Ext() != ".json" {
		return nil, input.UnrecognizedFileExtension
	}

	contents, err := file.Contents()
	if err != nil {
		return nil, err
	}

	var state models.State
	if err := json.Unmarshal(contents, &state); err != nil {
		return nil, input.FailedToParseInput
	}

	if state.InputType != "cloud_scan" {
		return nil, input.FailedToParseInput
	}

	return cloudScan{path: file.Path, state: state}, nil
}

func (cloudScanDetector) DetectDirectory(*input.Directory, input.DetectOptions) (input.IACConfiguration, error) {
	return nil, nil
}

func (c cloudScan) Type() *input.Type {
	return input.CloudScan
}

func (c cloudScan) ToState() models.State {
	return c.state
}

func (c cloudScan) LoadedFiles() []string {
	return []string{c.path}
}

func (cloudScan) Errors() []error {
	return nil
}

func (cloudScan) Location([]interface{}) ([]input.Location, error) {
	return nil, nil
}
