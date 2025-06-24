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

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLibFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/lib", 0755)
	afero.WriteFile(fsys, "existing/lib/relations.rego", []byte(relationsStub), 0644)

	expectedModule, err := ast.ParseModule("existing/lib/relations.rego", relationsStub)
	assert.NoError(t, err)

	testCases := []struct {
		name     string
		root     string
		expected *libDir
	}{
		{
			name: "from existing lib dir",
			root: "existing",
			expected: &libDir{
				Dir: ExistingDir("existing/lib"),
				relations: &relationsFile{
					module: expectedModule,
					lines:  3,
					File:   ExistingFile("existing/lib/relations.rego"),
				},
			},
		},
		{
			name: "non-existing lib dir",
			root: "empty",
			expected: &libDir{
				Dir:       NewDir("empty/lib"),
				relations: newRelationsFile(NewFile("empty/lib/relations.rego")),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l, err := libFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, l)
		})
	}
}

func TestLibWriteChanges(t *testing.T) {
	t.Run("should do nothing when dir exists", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		l := &libDir{
			Dir: ExistingDir("lib"),
			relations: &relationsFile{
				File: ExistingFile("lib/relations.rego"),
			},
		}
		err := l.WriteChanges(fsys)
		assert.NoError(t, err)
		exists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("should create dir and relations when dir does not exist", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		l := &libDir{
			Dir:       NewDir("lib"),
			relations: newRelationsFile(NewFile("lib/relations.rego")),
		}
		err := l.WriteChanges(fsys)
		assert.NoError(t, err)
		dirExists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.True(t, dirExists)
		relationsExists, err := afero.Exists(fsys, "lib/relations.rego")
		assert.NoError(t, err)
		assert.True(t, relationsExists)
	})
}
