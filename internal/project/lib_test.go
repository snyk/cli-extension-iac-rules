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
		relations, err := newRelationsFile(ExistingFile("lib/relations.rego"))
		assert.NoError(t, err)
		l := &libDir{
			Dir:       ExistingDir("lib"),
			relations: relations,
		}
		err = l.WriteChanges(fsys)
		assert.NoError(t, err)
		exists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("should create dir and relations when dir does not exist", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		relations, err := newRelationsFile(NewFile("lib/relations.rego"))
		assert.NoError(t, err)
		l := &libDir{
			Dir:       NewDir("lib"),
			relations: relations,
		}
		err = l.WriteChanges(fsys)
		assert.NoError(t, err)
		dirExists, err := afero.DirExists(fsys, "lib")
		assert.NoError(t, err)
		assert.True(t, dirExists)
		relationsExists, err := afero.Exists(fsys, "lib/relations")
		assert.NoError(t, err)
		assert.True(t, relationsExists)
	})
}
