package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSpecsFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/specs/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/specs/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/ignored.txt", []byte{}, 0644)
	testCases := []struct {
		name     string
		root     string
		expected *specsDir
	}{
		{
			name: "specs dir doesn't exist",
			root: "empty",
			expected: &specsDir{
				Dir:       NewDir("empty/specs"),
				ruleSpecs: map[string]*ruleSpecsDir{},
			},
		},
		{
			name: "existing specs dir",
			root: "existing",
			expected: &specsDir{
				Dir: ExistingDir("existing/specs"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: ExistingDir("existing/specs/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/specs/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/specs/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/specs/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/specs/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/specs/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td, err := specsFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, td)
		})
	}
}

func TestSpecsDirWriteChanges(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("new", 0755)
	fsys.MkdirAll("existing/specs/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/specs/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/specs/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	testCases := []struct {
		name string
		root string
		td   *specsDir
	}{
		{
			name: "new specs dir",
			root: "new",
			td: &specsDir{
				Dir: NewDir("new/specs"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: NewDir("new/specs/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    NewFile("new/specs/rules/TEST_001/inputs/infra.tf"),
								Expected: NewFile("new/specs/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: NewFile("new/specs/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    NewDir("new/specs/rules/TEST_001/inputs/invalid_ec2"),
								Expected: NewFile("new/specs/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
		{
			name: "existing specs dir",
			root: "existing",
			td: &specsDir{
				Dir: ExistingDir("existing/specs"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: ExistingDir("existing/specs/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/specs/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/specs/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/specs/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/specs/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/specs/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.td.WriteChanges(fsys)
			assert.NoError(t, err)
			output, err := specsFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.td, output)
		})
	}
}
