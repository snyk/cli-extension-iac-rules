// © 2023 Snyk Limited All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package project

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// ErrRuleDirAlreadyExists is returned when a rule already exists
var ErrRuleDirAlreadyExists = errors.New("rule directory already exists")

type rulesDir struct {
	*Dir
	rules map[string]*ruleDir
}

func (r *rulesDir) WriteChanges(fsys afero.Fs) error {
	if err := r.Dir.WriteChanges(fsys); err != nil {
		return err
	}
	for _, rule := range r.rules {
		if err := rule.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

func (r *rulesDir) addRule(ruleDirName string, regoFileName string, contents []byte) (string, error) {
	existing, exists := r.rules[ruleDirName]
	if exists {
		return "", fmt.Errorf("%w: %s", ErrRuleDirAlreadyExists, existing.Path())
	}
	path := filepath.Join(r.path, ruleDirName)
	r.rules[ruleDirName] = newRuleDir(path, regoFileName, contents)
	return r.rules[ruleDirName].files[regoFileName].Path(), nil
}

func (r *rulesDir) ruleDirNames() []string {
	var names []string
	for n := range r.rules {
		names = append(names, n)
	}
	return names
}

func rulesFromDir(fsys afero.Fs, root string) (*rulesDir, error) {
	path := filepath.Join(root, "rules")
	dir, err := DirFromPath(fsys, path)
	if err != nil {
		return nil, err
	}
	if !dir.Exists() {
		r := &rulesDir{
			Dir:   dir,
			rules: map[string]*ruleDir{},
		}
		return r, nil
	}
	entries, err := afero.ReadDir(fsys, path)
	if err != nil {
		return nil, err
	}
	rules := map[string]*ruleDir{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		r, err := ruleFromDir(fsys, path, name)
		if err != nil {
			return nil, err
		}
		rules[name] = r
	}
	r := &rulesDir{
		Dir:   dir,
		rules: rules,
	}
	return r, nil
}

type ruleDir struct {
	*Dir
	files map[string]FSNode
}

func (r *ruleDir) WriteChanges(fsys afero.Fs) error {
	if err := r.Dir.WriteChanges(fsys); err != nil {
		return err
	}
	for _, f := range r.files {
		if err := f.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

func ruleFromDir(fsys afero.Fs, parent string, name string) (*ruleDir, error) {
	path := filepath.Join(parent, name)
	entries, err := afero.ReadDir(fsys, path)
	if err != nil {
		return nil, readPathError(path, err)
	}
	files := map[string]FSNode{}
	for _, e := range entries {
		n := e.Name()
		files[n] = FSNodeFromFileInfo(path, e)
	}
	r := &ruleDir{
		Dir:   ExistingDir(path),
		files: files,
	}
	return r, nil
}

func newRuleDir(path string, regoFileName string, contents []byte) *ruleDir {
	regoFile := NewFile(filepath.Join(path, regoFileName))
	regoFile.UpdateContents(contents)
	return &ruleDir{
		Dir: NewDir(path),
		files: map[string]FSNode{
			regoFileName: regoFile,
		},
	}
}
