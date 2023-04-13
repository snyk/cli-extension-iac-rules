package project

import (
	"path/filepath"

	"github.com/spf13/afero"
)

type libDir struct {
	*Dir
}

func (l *libDir) WriteChanges(fsys afero.Fs) error {
	// Slightly different behavior, because users populate this directory
	// manually. We'll create a .gitkeep file so that this directory ends up
	// in version control.
	if !l.Exists() {
		if err := l.Dir.WriteChanges(fsys); err != nil {
			return err
		}

		gitkeep := NewFile(filepath.Join(l.Path(), ".gitkeep"))
		if err := gitkeep.WriteChanges(fsys); err != nil {
			return err
		}
	}

	return nil
}

func libFromDir(fsys afero.Fs, root string) (*libDir, error) {
	path := filepath.Join(root, "lib")
	dir, err := DirFromPath(fsys, path)
	if err != nil {
		return nil, err
	}
	return &libDir{Dir: dir}, nil
}
