package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var testRelationsFile = []byte(`package relations

import data.snyk

relations[info] {
	info := snyk.relation_from_fields(
		"aws_s3_bucket.logging",
		{"aws_s3_bucket": ["id", "bucket"]},
		{"aws_s3_bucket_logging": ["bucket"]},
	)
}
`)

var testRule = []byte(`package rules.TEST_001

input_type := "tf"
resource_type := "aws_s3_bucket"

metadata := {
	"id": "TEST-001",
	"severity": "high",
	"title": "S3 bucket has the word 'bucket' in its name",
	"description": "The word 'bucket' is redundant in a bucket name. We already know it's a bucket.",
	"product": [
			"iac",
			"cloud"
	]
}

deny[info] {
	contains(input.bucket, "bucket")
	info := {
		"resource": input
	}
}

`)

func TestProjectFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.MkdirAll("existing/lib", 0755)
	fsys.MkdirAll("existing/rules/TEST_001", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/inputs", 0755)
	fsys.MkdirAll("existing/spec/rules/TEST_001/expected", 0755)
	afero.WriteFile(fsys, "existing/manifest.json", []byte(`{"name":"Test"}`), 0644)
	afero.WriteFile(fsys, "existing/lib/relations.rego", testRelationsFile, 0644)
	afero.WriteFile(fsys, "existing/rules/TEST_001/main.rego", testRule, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
	afero.WriteFile(fsys, "existing/spec/rules/TEST_001/expected/infra.json", []byte{}, 0644)
	testCases := []struct {
		name              string
		root              string
		expectedManifest  Manifest
		expectedRules     []string
		expectedRuleSpecs []*RuleSpec
		expectedRelations []string
		expectedMetadata  map[string]RuleMetadata
	}{
		{
			name:              "project dir doesn't exist",
			root:              "nonexistent",
			expectedManifest:  Manifest{},
			expectedRules:     nil,
			expectedRuleSpecs: nil,
			expectedRelations: nil,
			expectedMetadata:  map[string]RuleMetadata{},
		},
		{
			name: "existing project dir",
			root: "existing",
			expectedManifest: Manifest{
				Name: "Test",
			},
			expectedRules: []string{"TEST_001"},
			expectedRuleSpecs: []*RuleSpec{
				{
					name:        "infra.tf",
					RuleDirName: "TEST_001",
					Input:       ExistingFile("existing/spec/rules/TEST_001/inputs/infra.tf"),
					Expected:    ExistingFile("existing/spec/rules/TEST_001/expected/infra.json"),
				},
			},
			expectedRelations: []string{"aws_s3_bucket.logging"},
			expectedMetadata: map[string]RuleMetadata{
				"TEST-001": {
					ID:          "TEST-001",
					Severity:    "high",
					Title:       "S3 bucket has the word 'bucket' in its name",
					Description: "The word 'bucket' is redundant in a bucket name. We already know it's a bucket.",
					Product:     []string{"iac", "cloud"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := FromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedManifest, p.Manifest())
			assert.Equal(t, tc.expectedRules, p.ListRules())
			assert.Equal(t, tc.expectedRuleSpecs, p.RuleSpecs())
			relations, err := p.RelationNames()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedRelations, relations)
			metadata, err := p.RuleMetadata()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedMetadata, metadata)
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
		path, err := p.AddRule("TEST_001", "main.rego", []byte{})
		assert.NoError(t, err)
		assert.Equal(t, "new/rules/TEST_001/main.rego", path)
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the new rule is listed
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, []string{"TEST_001"}, updated.ListRules())
	})

	t.Run("added rule spec", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Empty(t, p.RuleSpecs())

		// Add a test fixture and write the changes to disk
		p.AddRuleSpec("TEST_001", "infra.tf", []byte{})
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the new test fixture is
		// returned
		expected := []*RuleSpec{
			{
				name:        "infra.tf",
				RuleDirName: "TEST_001",
				Input:       ExistingFile("new/spec/rules/TEST_001/inputs/infra.tf"),
			},
		}
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, expected, updated.RuleSpecs())
	})

	t.Run("updated rule spec", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		fsys.MkdirAll("new/tests/rules/TEST_001/inputs", 0755)
		afero.WriteFile(fsys, "new/spec/rules/TEST_001/inputs/infra.tf", []byte{}, 0644)
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)

		// Update the test fixture with an expected output
		fixtures := p.RuleSpecs()
		assert.Equal(t, []*RuleSpec{
			{
				name:        "infra.tf",
				RuleDirName: "TEST_001",
				Input:       ExistingFile("new/spec/rules/TEST_001/inputs/infra.tf"),
			},
		}, fixtures)
		fixtures[0].UpdateExpected([]byte{})
		err = p.WriteChanges()
		assert.NoError(t, err)

		// Re-read the project from disk and assert that the test fixture we
		// updated in-place got written to disk when we called WriteChanges.
		expected := []*RuleSpec{
			{
				name:        "infra.tf",
				RuleDirName: "TEST_001",
				Input:       ExistingFile("new/spec/rules/TEST_001/inputs/infra.tf"),
				Expected:    ExistingFile("new/spec/rules/TEST_001/expected/infra.json"),
			},
		}
		updated, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		assert.Equal(t, expected, updated.RuleSpecs())
	})

	t.Run("added relation", func(t *testing.T) {
		// Initialize a new project
		fsys := afero.NewMemMapFs()
		p, err := FromDir(fsys, "new")
		assert.NoError(t, err)
		relations, err := p.RelationNames()
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
		relations, err = updated.RelationNames()
		assert.NoError(t, err)
		assert.Equal(t, []string{"aws_s3_bucket.logging"}, relations)
	})
}
