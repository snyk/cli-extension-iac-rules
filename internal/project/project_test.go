package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestProjectFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.MkdirAll("existing/lib", 0755)
	fsys.MkdirAll("existing/rules/TEST_001", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/inputs", 0755)
	fsys.MkdirAll("existing/tests/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/manifest.json", []byte(`{"name":"Test"}`), 0644)
	afero.WriteFile(fsys, "existing/lib/utils.rego", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/rules/TEST_001/main.rego", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/tests/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	testCases := []struct {
		name     string
		root     string
		expected *Project
	}{
		{
			name: "project dir doesn't exist",
			root: "nonexistent",
			expected: &Project{
				Dir: NewDir("nonexistent"),
				FS:  fsys,
				rulesDir: &rulesDir{
					Dir:   NewDir("nonexistent/rules"),
					rules: map[string]*ruleDir{},
				},
				libDir: &libDir{
					Dir: NewDir("nonexistent/lib"),
				},
				testsDir: &testsDir{
					Dir:       NewDir("nonexistent/tests"),
					ruleTests: map[string]*ruleTestDir{},
				},
				manifestFile: &manifestFile{
					File: NewFile("nonexistent/manifest.json"),
				},
			},
		},
		{
			name: "existing project dir",
			root: "existing",
			expected: &Project{
				Dir: ExistingDir("existing"),
				FS:  fsys,
				rulesDir: &rulesDir{
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
				libDir: &libDir{
					Dir: ExistingDir("existing/lib"),
				},
				testsDir: &testsDir{
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
							},
						},
					},
				},
				manifestFile: &manifestFile{
					File: ExistingFile("existing/manifest.json"),
					manifest: Manifest{
						Name: "Test",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := FromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, p)
		})
	}
}

func TestProjectWriteChanges(t *testing.T) {
	t.Run("updated manifest", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, Manifest{}, p.Manifest())

		// Update the manifest and write the changes to disk
		manifest := Manifest{
			Name: "Test",
		}
		p.UpdateManifest(manifest)
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the manifest was
		// updated as expected.
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, manifest, updated.Manifest())
	})

	t.Run("added rule", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Empty(t, p.ListRules())

		// Add a rule and write the changes to disk
		err = p.AddRule("TEST_001", "main.rego", []byte{})
		assert.NoError(t, err)
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the new rule is listed
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, []string{"TEST_001"}, updated.ListRules())
	})

	t.Run("added rule test fixture", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Empty(t, p.RuleTestFixtures())

		// Add a test fixture and write the changes to disk
		p.AddRuleTestFixture("TEST_001", "infra.tf", []byte{})
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the new test fixture is
		// returned
		expected := []*RuleTestFixture{
			{
				name:  "infra.tf",
				Input: ExistingFile("new/tests/rules/TEST_001/inputs/infra.tf"),
			},
		}
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, expected, updated.RuleTestFixtures())
	})

	t.Run("updated rule test fixture", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		fsys.MkdirAll("new/tests/rules/TEST_001/inputs", 0755)
		afero.WriteFile(fsys, "new/tests/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)

		// Update the test fixture with an expected output
		fixtures := p.RuleTestFixtures()
		assert.Equal(t, []*RuleTestFixture{
			{
				name:  "infra.tf",
				Input: ExistingFile("new/tests/rules/TEST_001/inputs/infra.tf"),
			},
		}, fixtures)
		fixtures[0].UpdateExpected([]byte{})
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the test fixture we
		// updated in-place got written to disk when we called WriteChanges.
		expected := []*RuleTestFixture{
			{
				name:     "infra.tf",
				Input:    ExistingFile("new/tests/rules/TEST_001/inputs/infra.tf"),
				Expected: ExistingFile("new/tests/rules/TEST_001/expected/infra.json"),
			},
		}
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, expected, updated.RuleTestFixtures())
	})

	t.Run("added relation", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		relations, err := p.ListRelations()
		assert.NoError(t, err)
		assert.Empty(t, relations)

		// Add a relation and write the changes to disk
		err = p.AddRelation(`relation[info] {
			info := snyk.relation_from_fields(
                "aws_s3_bucket.logging",
                {"aws_s3_bucket": ["id", "bucket"]},
                {"aws_s3_bucket_logging": ["bucket"]},
        	)
		}`)
		assert.NoError(t, err)
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the new rule is listed
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		relations, err = updated.ListRelations()
		assert.NoError(t, err)
		assert.Equal(t, []string{"aws_s3_bucket.logging"}, relations)
	})
}
