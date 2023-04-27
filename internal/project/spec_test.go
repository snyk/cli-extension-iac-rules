package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSpecFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/ignored.txt", []byte{}, 0644)
	testCases := []struct {
		name     string
		root     string
		expected *specDir
	}{
		{
			name: "spec dir doesn't exist",
			root: "empty",
			expected: &specDir{
				Dir:       NewDir("empty/spec"),
				ruleSpecs: map[string]*ruleSpecsDir{},
			},
		},
		{
			name: "existing spec dir",
			root: "existing",
			expected: &specDir{
				Dir: ExistingDir("existing/spec"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: ExistingDir("existing/spec/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/spec/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/spec/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/spec/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/spec/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/spec/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td, err := specFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, td)
		})
	}
}

func TestSpecDirWriteChanges(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("new", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	testCases := []struct {
		name string
		root string
		td   *specDir
	}{
		{
			name: "new spec dir",
			root: "new",
			td: &specDir{
				Dir: NewDir("new/spec"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: NewDir("new/spec/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    NewFile("new/spec/rules/TEST_001/inputs/infra.tf"),
								Expected: NewFile("new/spec/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: NewFile("new/spec/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    NewDir("new/spec/rules/TEST_001/inputs/invalid_ec2"),
								Expected: NewFile("new/spec/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
		{
			name: "existing spec dir",
			root: "existing",
			td: &specDir{
				Dir: ExistingDir("existing/spec"),
				ruleSpecs: map[string]*ruleSpecsDir{
					"TEST_001": {
						Dir: ExistingDir("existing/spec/rules/TEST_001"),
						fixtures: map[string]*RuleSpec{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/spec/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/spec/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/spec/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/spec/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/spec/rules/TEST_001/expected/invalid_ec2.json"),
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
			output, err := specFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.td, output)
		})
	}
}
