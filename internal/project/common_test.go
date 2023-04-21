package project

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFileWriteChanges(t *testing.T) {
	t.Run("should create a file", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		file := NewFile("test")
		err := file.WriteChanges(fsys)
		assert.NoError(t, err)
		stat, err := fsys.Stat("test")
		assert.NoError(t, err)
		assert.False(t, stat.IsDir())
	})
	t.Run("shouldn't try to recreate file when exists is true", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		file := ExistingFile("test")
		err := file.WriteChanges(fsys)
		assert.NoError(t, err)
		exists, err := afero.Exists(fsys, "test")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("should recreate file when content is set", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		file := ExistingFile("test")
		file.UpdateContents([]byte{})
		err := file.WriteChanges(fsys)
		assert.NoError(t, err)
		stat, err := fsys.Stat("test")
		assert.NoError(t, err)
		assert.False(t, stat.IsDir())
	})
	t.Run("should produce an error when create file fails", func(t *testing.T) {
		fsys := afero.NewReadOnlyFs(afero.NewMemMapFs())
		file := NewFile("test")
		err := file.WriteChanges(fsys)
		assert.ErrorIs(t, err, ErrFailedToCreateFile)
	})
}

func TestDirWriteChanges(t *testing.T) {
	t.Run("should create a directory", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		dir := NewDir("test")
		err := dir.WriteChanges(fsys)
		assert.NoError(t, err)
		stat, err := fsys.Stat("test")
		assert.NoError(t, err)
		assert.True(t, stat.IsDir())
	})
	t.Run("shouldn't try to recreate directory when exists is true", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		dir := ExistingDir("test")
		err := dir.WriteChanges(fsys)
		assert.NoError(t, err)
		exists, err := afero.DirExists(fsys, "test")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
	t.Run("should produce an error when create directory fails", func(t *testing.T) {
		fsys := afero.NewReadOnlyFs(afero.NewMemMapFs())
		dir := NewDir("test")
		err := dir.WriteChanges(fsys)
		assert.ErrorIs(t, err, ErrFailedToCreateDir)
	})
}

func TestSafeNames(t *testing.T) {
	t.Run("safeFileName", func(t *testing.T) {
		for _, tc := range []struct {
			input       string
			expected    string
			expectedErr error
		}{
			{
				input:    "infra.tf",
				expected: "infra.tf",
			},
			{
				input:    "invalid_ec2",
				expected: "invalid_ec2",
			},
			{
				input:    "some invalid-file.tf",
				expected: "some_invalid_file.tf",
			},
			{
				input:       "01-test.json",
				expectedErr: ErrInvalidIdentifier,
			},
		} {
			t.Run(tc.input, func(t *testing.T) {
				output, err := safeFilename(tc.input)
				assert.Equal(t, tc.expected, output)
				assert.ErrorIs(t, err, tc.expectedErr)
			})
		}
	})

	t.Run("safePackageName", func(t *testing.T) {
		for _, tc := range []struct {
			input       string
			expected    string
			expectedErr error
		}{
			{
				input:    "SNYK-CC-00001",
				expected: "SNYK_CC_00001",
			},
			{
				input:    "TEST_001",
				expected: "TEST_001",
			},
			{
				input:    "test 😉01",
				expected: "test__01",
			},
			{
				input:    "test 😉01. foo",
				expected: "test__01__foo",
			},
			{
				input:       "01-test",
				expectedErr: ErrInvalidIdentifier,
			},
		} {
			t.Run(tc.input, func(t *testing.T) {
				output, err := SafePackageName(tc.input)
				assert.Equal(t, tc.expected, output)
				assert.ErrorIs(t, err, tc.expectedErr)
			})
		}
	})
}
