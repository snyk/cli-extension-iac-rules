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
	name        string
	RuleDirName string
	Input       FSNode
	Expected    *File
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
		f.Expected = NewFile(f.ExpectedPath())
	}
	f.Expected.UpdateContents(contents)
}

func (f *RuleSpec) ExpectedPath() string {
	noExt := strings.TrimSuffix(f.name, filepath.Ext(f.name))
	parent := filepath.Dir(f.Input.Path())
	expectedName := fmt.Sprintf("%s.json", noExt)
	return filepath.Join(parent, "..", "expected", expectedName)
}

func ruleSpecFromFileInfo(fsys afero.Fs, parent string, info fs.FileInfo, ruleDirName string) (*RuleSpec, error) {
	fixture := &RuleSpec{
		name:        info.Name(),
		RuleDirName: ruleDirName,
		Input:       FSNodeFromFileInfo(parent, info),
	}
	expectedPath := fixture.ExpectedPath()
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

type specDir struct {
	*Dir
	ruleSpecs map[string]*ruleSpecsDir
}

func (t *specDir) WriteChanges(fsys afero.Fs) error {
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

func (t *specDir) fixtures() []*RuleSpec {
	var fixtures []*RuleSpec
	for _, r := range t.ruleSpecs {
		for _, f := range r.fixtures {
			fixtures = append(fixtures, f)
		}
	}
	return fixtures
}

func (t *specDir) addRuleSpecsDir(ruleDirName string) *ruleSpecsDir {
	t.ruleSpecs[ruleDirName] = &ruleSpecsDir{
		Dir:      NewDir(filepath.Join(t.Path(), "rules", ruleDirName)),
		fixtures: map[string]*RuleSpec{},
	}
	return t.ruleSpecs[ruleDirName]
}

func (t *specDir) addRuleSpec(ruleDirName string, name string, contents []byte) (string, error) {
	rt, ok := t.ruleSpecs[ruleDirName]
	if !ok {
		rt = t.addRuleSpecsDir(ruleDirName)
	}
	return rt.addFixture(name, contents)
}

func specFromDir(fsys afero.Fs, root string) (*specDir, error) {
	specPath := filepath.Join(root, "spec")
	dir, err := DirFromPath(fsys, specPath)
	if err != nil {
		return nil, err
	}
	if !dir.Exists() {
		t := &specDir{
			Dir:       dir,
			ruleSpecs: map[string]*ruleSpecsDir{},
		}
		return t, nil
	}
	entries, err := afero.ReadDir(fsys, specPath)
	if err != nil {
		return nil, readPathError(specPath, err)
	}
	ruleSpecs := map[string]*ruleSpecsDir{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "rules" {
			rulesDir := filepath.Join(specPath, e.Name())
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
	t := &specDir{
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

func (t *ruleSpecsDir) addFixture(name string, contents []byte) (string, error) {
	f, exists := t.fixtures[name]
	if exists {
		return "", fmt.Errorf("%w: %s", ErrRuleSpecAlreadyExists, f.Input.Path())
	}
	input := NewFile(filepath.Join(t.path, "inputs", name))
	input.UpdateContents(contents)
	t.fixtures[name] = &RuleSpec{
		name:  name,
		Input: input,
	}
	return input.Path(), nil
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
				f, err := ruleSpecFromFileInfo(fsys, inputsDir, e, name)
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
