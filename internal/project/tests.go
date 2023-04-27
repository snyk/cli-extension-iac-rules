package project

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var ErrRuleSpecAlreadyExists = errors.New("rule spec already exists")

// RuleSpec represents an input file or directory and an expected output
// file.
type RuleSpec struct {
	name     string
	Input    FSNode
	Expected *File
}

// WriteChanges persists any changes to this fixture to disk.
func (f *RuleSpec) WriteChanges(fsys afero.Fs) error {
	if err := f.Input.WriteChanges(fsys); err != nil {
		return err
	}
	if f.Expected != nil {
		if err := f.Expected.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

// UpdateExpected updates the expected output file for this fixture.
func (f *RuleSpec) UpdateExpected(contents []byte) {
	if f.Expected == nil {
		f.Expected = NewFile(f.expectedPath())
	}
	f.Expected.UpdateContents(contents)
}

func (f *RuleSpec) expectedPath() string {
	noExt := strings.TrimSuffix(f.name, filepath.Ext(f.name))
	parent := filepath.Dir(f.Input.Path())
	expectedName := fmt.Sprintf("%s.json", noExt)
	return filepath.Join(parent, "..", "expected", expectedName)
}

func ruleSpecFromFileInfo(fsys afero.Fs, parent string, info fs.FileInfo) (*RuleSpec, error) {
	fixture := &RuleSpec{
		name:  info.Name(),
		Input: FSNodeFromFileInfo(parent, info),
	}
	expectedPath := fixture.expectedPath()
	expectedFile, err := FileFromPath(fsys, expectedPath)
	if err != nil {
		return nil, err
	}
	if expectedFile.Exists() {
		// Only want to set expected if it already exists so that we don't
		// create empty JSON files.
		fixture.Expected = expectedFile
	}
	return fixture, nil
}

type specsDir struct {
	*Dir
	ruleSpecs map[string]*ruleSpecsDir
}

func (t *specsDir) WriteChanges(fsys afero.Fs) error {
	if err := t.Dir.WriteChanges(fsys); err != nil {
		return err
	}
	for _, rt := range t.ruleSpecs {
		if err := rt.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

func (t *specsDir) fixtures() []*RuleSpec {
	var fixtures []*RuleSpec
	for _, r := range t.ruleSpecs {
		for _, f := range r.fixtures {
			fixtures = append(fixtures, f)
		}
	}
	return fixtures
}

func (t *specsDir) addRuleSpecsDir(ruleDirName string) *ruleSpecsDir {
	t.ruleSpecs[ruleDirName] = &ruleSpecsDir{
		Dir:      NewDir(filepath.Join(t.Path(), "rules", ruleDirName)),
		fixtures: map[string]*RuleSpec{},
	}
	return t.ruleSpecs[ruleDirName]
}

func (t *specsDir) addRuleSpec(ruleDirName string, name string, contents []byte) error {
	rt, ok := t.ruleSpecs[ruleDirName]
	if !ok {
		rt = t.addRuleSpecsDir(ruleDirName)
	}
	return rt.addFixture(name, contents)
}

func specsFromDir(fsys afero.Fs, root string) (*specsDir, error) {
	specsPath := filepath.Join(root, "specs")
	dir, err := DirFromPath(fsys, specsPath)
	if err != nil {
		return nil, err
	}
	if !dir.Exists() {
		t := &specsDir{
			Dir:       dir,
			ruleSpecs: map[string]*ruleSpecsDir{},
		}
		return t, nil
	}
	entries, err := afero.ReadDir(fsys, specsPath)
	if err != nil {
		return nil, readPathError(specsPath, err)
	}
	ruleSpecs := map[string]*ruleSpecsDir{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "rules" {
			rulesDir := filepath.Join(specsPath, e.Name())
			entries, err := afero.ReadDir(fsys, rulesDir)
			if err != nil {
				return nil, readPathError(rulesDir, err)
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				rt, err := ruleSpecsFromDir(fsys, rulesDir, name)
				if err != nil {
					return nil, err
				}
				ruleSpecs[name] = rt
			}
		}
	}
	t := &specsDir{
		Dir:       dir,
		ruleSpecs: ruleSpecs,
	}
	return t, nil
}

type ruleSpecsDir struct {
	*Dir
	fixtures map[string]*RuleSpec
}

func (t *ruleSpecsDir) WriteChanges(fsys afero.Fs) error {
	if err := t.Dir.WriteChanges(fsys); err != nil {
		return err
	}

	for _, f := range t.fixtures {
		if err := f.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

func (t *ruleSpecsDir) addFixture(name string, contents []byte) error {
	f, exists := t.fixtures[name]
	if exists {
		return fmt.Errorf("%w: %s", ErrRuleSpecAlreadyExists, f.Input.Path())
	}
	input := NewFile(filepath.Join(t.path, "inputs", name))
	input.UpdateContents(contents)
	t.fixtures[name] = &RuleSpec{
		name:  name,
		Input: input,
	}
	return nil
}

func ruleSpecsFromDir(fsys afero.Fs, parent string, name string) (*ruleSpecsDir, error) {
	path := filepath.Join(parent, name)
	entries, err := afero.ReadDir(fsys, path)
	if err != nil {
		return nil, readPathError(path, err)
	}
	fixtures := map[string]*RuleSpec{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "inputs" {
			inputsDir := filepath.Join(path, e.Name())
			entries, err := afero.ReadDir(fsys, inputsDir)
			if err != nil {
				return nil, readPathError(inputsDir, err)
			}
			for _, e := range entries {
				f, err := ruleSpecFromFileInfo(fsys, inputsDir, e)
				if err != nil {
					return nil, err
				}
				fixtures[f.name] = f
			}
		}
	}
	t := &ruleSpecsDir{
		Dir:      ExistingDir(path),
		fixtures: fixtures,
	}
	return t, nil
}
