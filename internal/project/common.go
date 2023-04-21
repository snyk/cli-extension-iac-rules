package project

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

const directoryPermission = 0755
const filePermission = 0644

// ErrFailedToCreateDir is returned when we were unable to create a directory
var ErrFailedToCreateDir = errors.New("failed to create directory")

// ErrFailedToCreateFile is returned when we were unable to create a file
var ErrFailedToCreateFile = errors.New("failed to write to file")

// ErrFailedToReadPath is returned when we encountered a filesystem error while
// reading a path.
var ErrFailedToReadPath = errors.New("failed to read path")

// ErrUnexpectedType is returned when we expected a directory and found a file
// or vice versa.
var ErrUnexpectedType = errors.New("unexpected file type")

// ErrInvalidIdentifier is returned when an identifier does not satisfy some
// constraint.
var ErrInvalidIdentifier = errors.New("invalid identifier")

func pathError(path string, outer, inner error) error {
	return fmt.Errorf("%w %s: %s", outer, path, inner)
}

func readPathError(path string, err error) error {
	return pathError(path, ErrFailedToReadPath, err)
}

// FSNode defines the base set of operations for both files and directories.
type FSNode interface {
	// Path returns the path for this node.
	Path() string
	// Exists returns whether or not this node exists on disk.
	Exists() bool
	// IsDir returns whether or not this node is a directory.
	IsDir() bool
	// WriteChanges persists any changes to this node back to disk.
	WriteChanges(fsys afero.Fs) error
}

// FSNodeFromFileInfo returns an FSNode for the given fs.FileInfo object in the
// parent directory.
func FSNodeFromFileInfo(parent string, i fs.FileInfo) FSNode {
	path := filepath.Join(parent, i.Name())
	if i.IsDir() {
		return ExistingDir(path)
	}
	return ExistingFile(path)
}

// File represents a file on disk.
type File struct {
	path            string
	exists          bool
	dirty           bool
	pendingContents []byte
}

// NewFile returns a File object that represents a file that does not exist yet.
func NewFile(path string) *File {
	return &File{
		path:   path,
		exists: false,
	}
}

// ExistingFile returns a File object that represents an existing file on disk.
func ExistingFile(path string) *File {
	return &File{
		path:   path,
		exists: true,
	}
}

// FileFromPath returns a File for the given path whether it exists or not.
func FileFromPath(fsys afero.Fs, path string) (*File, error) {
	info, err := fsys.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return NewFile(path), nil
		} else {
			return nil, readPathError(path, err)
		}
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%w: %s is a directory", ErrUnexpectedType, path)
	}
	return ExistingFile(path), nil
}

// Path returns the path of this File object.
func (f *File) Path() string {
	return f.path
}

// Exists returns whether the represented file exists or not.
func (f *File) Exists() bool {
	return f.exists
}

// IsDir always returns false.
func (f *File) IsDir() bool {
	return false
}

// UpdateContents will stage changes to this file that will be persisted when
// WriteChanges is called.
func (f *File) UpdateContents(b []byte) {
	f.pendingContents = b
	f.dirty = true
}

// WriteChanges persists any changes to this file to disk.
func (f *File) WriteChanges(fsys afero.Fs) error {
	if f.exists && !f.dirty {
		return nil
	}
	parent := filepath.Dir(f.path)
	if err := fsys.MkdirAll(parent, directoryPermission); err != nil {
		return pathError(parent, ErrFailedToCreateDir, err)
	}
	if err := afero.WriteFile(fsys, f.path, f.pendingContents, filePermission); err != nil {
		return pathError(f.path, ErrFailedToCreateFile, err)
	}
	f.exists = true
	f.dirty = false
	f.pendingContents = nil
	return nil
}

// Dir represents a directory on disk.
type Dir struct {
	path   string
	exists bool
}

// NewDir returns a Dir object that represents a directory that does not exist
// yet.
func NewDir(path string) *Dir {
	return &Dir{
		path:   path,
		exists: false,
	}
}

// ExistingDir returns a Dir object that represents an existing directory on
// disk.
func ExistingDir(path string) *Dir {
	return &Dir{
		path:   path,
		exists: true,
	}
}

// DirFromPath returns a Dir for the given path whether it exists or not.
func DirFromPath(fsys afero.Fs, path string) (*Dir, error) {
	info, err := fsys.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return NewDir(path), nil
		} else {
			return nil, readPathError(path, err)
		}
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s is a file", ErrUnexpectedType, path)
	}
	return ExistingDir(path), nil
}

// Path returns the path of this Dir object.
func (d *Dir) Path() string {
	return d.path
}

// Exists returns whether the represented directory exists or not.
func (d *Dir) Exists() bool {
	return d.exists
}

// IsDir always returns true.
func (d *Dir) IsDir() bool {
	return true
}

// WriteChanges will create the directory on disk if it does not already exist.
func (d *Dir) WriteChanges(fsys afero.Fs) error {
	if d.exists {
		return nil
	}
	if err := fsys.MkdirAll(d.path, directoryPermission); err != nil {
		return pathError(d.path, ErrFailedToCreateDir, err)
	}
	d.exists = true
	return nil
}

var replaceCharsRegex = regexp.MustCompile(`[^0-9A-Za-z_.]`)
var validIdentifier = regexp.MustCompile(`^[[:alpha:]]`)

func safeFilename(s string) (string, error) {
	if !validIdentifier.MatchString(s) {
		return "", fmt.Errorf("%w %s: should start with a letter", ErrInvalidIdentifier, s)
	}
	return replaceCharsRegex.ReplaceAllString(s, "_"), nil
}

func SafePackageName(s string) (string, error) {
	safe, err := safeFilename(s)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(safe, ".", "_"), nil
}
