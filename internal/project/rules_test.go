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

package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRulesFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/rules/TEST_001", 0755)
	afero.WriteFile(fsys, "existing/rules/TEST_001/main.rego", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/rules/ignored.txt", []byte{}, 0644)
	testCases := []struct {
		name     string
		root     string
		expected *rulesDir
	}{
		{
			name: "rules dir doesn't exist",
			root: "empty",
			expected: &rulesDir{
				Dir:   NewDir("empty/rules"),
				rules: map[string]*ruleDir{},
			},
		},
		{
			name: "existing rules dir",
			root: "existing",
			expected: &rulesDir{
				Dir: ExistingDir("existing/rules"),
				rules: map[string]*ruleDir{
					"TEST_001": {
						Dir: ExistingDir("existing/rules/TEST_001"),
						files: map[string]FSNode{
							"main.rego": ExistingFile("existing/rules/TEST_001/main.rego"),
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := rulesFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, r)
		})
	}
}

func TestRulesDirCreate(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("new", 0755)
	fsys.MkdirAll("existing/rules/TEST_001", 0755)
	afero.WriteFile(fsys, "existing/rules/TEST_001/main.rego", []byte{}, 0644)
	testCases := []struct {
		name string
		root string
		r    *rulesDir
	}{
		{
			name: "new rules dir",
			root: "new",
			r: &rulesDir{
				Dir: NewDir("new/rules"),
				rules: map[string]*ruleDir{
					"TEST_001": {
						Dir: NewDir("new/rules/TEST_001"),
						files: map[string]FSNode{
							"main.rego": NewFile("new/rules/TEST_001/main.rego"),
						},
					},
				},
			},
		},
		{
			name: "existing rules dir",
			root: "existing",
			r: &rulesDir{
				Dir: ExistingDir("existing/rules"),
				rules: map[string]*ruleDir{
					"TEST_001": {
						Dir: ExistingDir("existing/rules/TEST_001"),
						files: map[string]FSNode{
							"main.rego": ExistingFile("existing/rules/TEST_001/main.rego"),
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.r.WriteChanges(fsys)
			assert.NoError(t, err)
			output, err := rulesFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.r, output)
		})
	}
}
