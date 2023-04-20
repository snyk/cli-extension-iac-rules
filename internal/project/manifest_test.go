package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestManifestFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.Mkdir("existing", 0755)
	fsys.Mkdir("error", 0755)
	afero.WriteFile(fsys, "existing/manifest.json", []byte(`{"name": "test"}`), 0644)
	afero.WriteFile(fsys, "error/manifest.json", []byte(`[]`), 0644)

	testCases := []struct {
		name          string
		root          string
		expected      *manifestFile
		expectedError error
	}{
		{
			name: "from existing manifest file",
			root: "existing",
			expected: &manifestFile{
				File: ExistingFile("existing/manifest.json"),
				manifest: Manifest{
					Name: "test",
				},
			},
		},
		{
			name: "non-existing manifest file",
			root: "empty",
			expected: &manifestFile{
				File:     NewFile("empty/manifest.json"),
				manifest: Manifest{},
			},
		},
		{
			name:          "invalid manifest file",
			root:          "error",
			expectedError: ErrFailedToUnmarshalManifest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := manifestFromDir(fsys, tc.root)
			if tc.expectedError != nil {
				assert.Nil(t, m)
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, m)
			}
		})
	}
}

func TestWriteChanges(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("new", 0755)
	fsys.Mkdir("existing", 0755)
	afero.WriteFile(fsys, "existing/manifest.json", []byte(`{"name": "test"}`), 0644)
	testCases := []struct {
		name string
		root string
		m    *manifestFile
	}{
		{
			name: "new manifest file",
			root: "new",
			m: &manifestFile{
				File: NewFile("new/manifest.json"),
				manifest: Manifest{
					Name: "test",
				},
			},
		},
		{
			name: "update existing manifest file",
			root: "existing",
			m: &manifestFile{
				File: ExistingFile("existing/manifest.json"),
				manifest: Manifest{
					Name: "updated",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.m.WriteChanges(fsys)
			assert.NoError(t, err)
			output, err := manifestFromDir(fsys, tc.root)
			assert.NoError(t, err)
			assert.Equal(t, tc.m, output)
		})
	}
}
