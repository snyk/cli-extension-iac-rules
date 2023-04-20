package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestTestsFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/ignored.txt", []byte{}, 0644)
	testCases := []struct {
		name     string
		root     string
		expected *testsDir
	}{
		{
			name: "tests dir doesn't exist",
			root: "empty",
			expected: &testsDir{
				Dir:       NewDir("empty/tests"),
				ruleTests: map[string]*ruleTestDir{},
			},
		},
		{
			name: "existing tests dir",
			root: "existing",
			expected: &testsDir{
				Dir: ExistingDir("existing/tests"),
				ruleTests: map[string]*ruleTestDir{
					"TEST_001": {
						Dir: ExistingDir("existing/tests/rules/TEST_001"),
						fixtures: map[string]*RuleTestFixture{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/tests/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/tests/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/tests/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/tests/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/tests/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			td, err := testsFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, td)
		})
	}
}

func TestTestsDirWriteChanges(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("new", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/inputs/invalid_ec2", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/no_expected.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/invalid_ec2/main.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/invalid_ec2/module.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/expected/invalid_ec2.json", []byte{}, 0644)
	testCases := []struct {
		name string
		root string
		td   *testsDir
	}{
		{
			name: "new tests dir",
			root: "new",
			td: &testsDir{
				Dir: NewDir("new/tests"),
				ruleTests: map[string]*ruleTestDir{
					"TEST_001": {
						Dir: NewDir("new/tests/rules/TEST_001"),
						fixtures: map[string]*RuleTestFixture{
							"infra.tf": {
								name:     "infra.tf",
								Input:    NewFile("new/tests/rules/TEST_001/inputs/infra.tf"),
								Expected: NewFile("new/tests/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: NewFile("new/tests/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    NewDir("new/tests/rules/TEST_001/inputs/invalid_ec2"),
								Expected: NewFile("new/tests/rules/TEST_001/expected/invalid_ec2.json"),
							},
						},
					},
				},
			},
		},
		{
			name: "existing tests dir",
			root: "existing",
			td: &testsDir{
				Dir: ExistingDir("existing/tests"),
				ruleTests: map[string]*ruleTestDir{
					"TEST_001": {
						Dir: ExistingDir("existing/tests/rules/TEST_001"),
						fixtures: map[string]*RuleTestFixture{
							"infra.tf": {
								name:     "infra.tf",
								Input:    ExistingFile("existing/tests/rules/TEST_001/inputs/infra.tf"),
								Expected: ExistingFile("existing/tests/rules/TEST_001/expected/infra.json"),
							},
							"no_expected.tf": {
								name:  "no_expected.tf",
								Input: ExistingFile("existing/tests/rules/TEST_001/inputs/no_expected.tf"),
							},
							"invalid_ec2": {
								name:     "invalid_ec2",
								Input:    ExistingDir("existing/tests/rules/TEST_001/inputs/invalid_ec2"),
								Expected: ExistingFile("existing/tests/rules/TEST_001/expected/invalid_ec2.json"),
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
			output, err := testsFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.td, output)
		})
	}
}
