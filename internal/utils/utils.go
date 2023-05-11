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

package utils

import (
	"fmt"

	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/models"
	"github.com/spf13/afero"
)

type SingleInput struct {
	State  models.State
	Loader input.Loader // Can be used to call AddSourceLocs
}

func LoadSingleInput(path string) (*SingleInput, error) {
	detector, err := input.DetectorByInputTypes(input.Types{input.Auto})
	if err != nil {
		return nil, err
	}
	detector = input.NewMultiDetector(cloudScanDetector{}, detector)
	loader := input.NewLoader(detector)
	fsys := afero.OsFs{}
	detectable, err := input.NewDetectable(fsys, path)
	if err != nil {
		return nil, err
	}
	_, err = loader.Load(detectable, input.DetectOptions{})
	if err != nil {
		return nil, err
	}
	states := loader.ToStates()
	if len(states) != 1 {
		return nil, fmt.Errorf("internal error: expected a single input but got %d", len(states))
	}
	return &SingleInput{State: states[0], Loader: loader}, nil
}
