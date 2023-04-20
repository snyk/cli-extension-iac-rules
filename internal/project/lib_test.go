package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLibFromDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	fsys.Mkdir("empty", 0755)
	fsys.MkdirAll("existing/lib", 0755)

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
			},
		},
		{
			name: "non-existing lib dir",
			root: "empty",
			expected: &libDir{
				Dir: NewDir("empty/lib"),
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
		}
		err := l.WriteChanges(fsys)
		assert.NoError(t, err)
		exists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("should create dir and gitkeep when dir does not exist", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		l := &libDir{
			Dir: NewDir("lib"),
		}
		err := l.WriteChanges(fsys)
		assert.NoError(t, err)
		dirExists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.True(t, dirExists)
		gitkeepExists, err := afero.Exists(fsys, "lib/.gitkeep")
		assert.NoError(t, err)
		assert.True(t, gitkeepExists)
	})
}
