package project

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var ErrRuleTestFixtureAlreadyExists = errors.New("rule test fixture already exists")

// RuleTestFixture represents an input file or directory and an expected output
// file.
type RuleTestFixture struct {
	name     string
	Input    FSNode
	Expected *File
}

// WriteChanges persists any changes to this fixture to disk.
func (f *RuleTestFixture) WriteChanges(fsys afero.Fs) error {
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
func (f *RuleTestFixture) UpdateExpected(contents []byte) {
	if f.Expected == nil {
		f.Expected = NewFile(f.expectedPath())
	}
	f.Expected.UpdateContents(contents)
}

func (f *RuleTestFixture) expectedPath() string {
	noExt := strings.TrimSuffix(f.name, filepath.Ext(f.name))
	parent := filepath.Dir(f.Input.Path())
	expectedName := fmt.Sprintf("%s.json", noExt)
	return filepath.Join(parent, "..", "expected", expectedName)
}

func ruleTestFixtureFromFileInfo(fsys afero.Fs, parent string, info fs.FileInfo) (*RuleTestFixture, error) {
	fixture := &RuleTestFixture{
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

type testsDir struct {
	*Dir
	ruleTests map[string]*ruleTestDir
}

func (t *testsDir) WriteChanges(fsys afero.Fs) error {
	if err := t.Dir.WriteChanges(fsys); err != nil {
		return err
	}
	for _, rt := range t.ruleTests {
		if err := rt.WriteChanges(fsys); err != nil {
			return err
		}
	}
	return nil
}

func (t *testsDir) fixtures() []*RuleTestFixture {
	var fixtures []*RuleTestFixture
	for _, r := range t.ruleTests {
		for _, f := range r.fixtures {
			fixtures = append(fixtures, f)
		}
	}
	return fixtures
}

func (t *testsDir) addRuleTestDir(ruleDirName string) *ruleTestDir {
	t.ruleTests[ruleDirName] = &ruleTestDir{
		Dir:      NewDir(filepath.Join(t.Path(), "rules", ruleDirName)),
		fixtures: map[string]*RuleTestFixture{},
	}
	return t.ruleTests[ruleDirName]
}

func (t *testsDir) addRuleTestFixture(ruleDirName string, name string, contents []byte) error {
	rt, ok := t.ruleTests[ruleDirName]
	if !ok {
		rt = t.addRuleTestDir(ruleDirName)
	}
	return rt.addFixture(name, contents)
}

func testsFromDir(fsys afero.Fs, root string) (*testsDir, error) {
	testDir := filepath.Join(root, "tests")
	dir, err := DirFromPath(fsys, testDir)
	if err != nil {
		return nil, err
	}
	if !dir.Exists() {
		t := &testsDir{
			Dir:       dir,
			ruleTests: map[string]*ruleTestDir{},
		}
		return t, nil
	}
	entries, err := afero.ReadDir(fsys, testDir)
	if err != nil {
		return nil, readPathError(testDir, err)
	}
	ruleTests := map[string]*ruleTestDir{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "rules" {
			rulesDir := filepath.Join(testDir, e.Name())
			entries, err := afero.ReadDir(fsys, rulesDir)
			if err != nil {
				return nil, readPathError(rulesDir, err)
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				rt, err := ruleTestFromDir(fsys, rulesDir, name)
				if err != nil {
					return nil, err
				}
				ruleTests[name] = rt
			}
		}
	}
	t := &testsDir{
		Dir:       dir,
		ruleTests: ruleTests,
	}
	return t, nil
}

type ruleTestDir struct {
	*Dir
	fixtures map[string]*RuleTestFixture
}

func (t *ruleTestDir) WriteChanges(fsys afero.Fs) error {
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

func (t *ruleTestDir) addFixture(name string, contents []byte) error {
	f, exists := t.fixtures[name]
	if exists {
		return fmt.Errorf("%w: %s", ErrRuleTestFixtureAlreadyExists, f.Input.Path())
	}
	input := NewFile(filepath.Join(t.path, "inputs", name))
	input.UpdateContents(contents)
	t.fixtures[name] = &RuleTestFixture{
		name:  name,
		Input: input,
	}
	return nil
}

func ruleTestFromDir(fsys afero.Fs, parent string, name string) (*ruleTestDir, error) {
	path := filepath.Join(parent, name)
	entries, err := afero.ReadDir(fsys, path)
	if err != nil {
		return nil, readPathError(path, err)
	}
	fixtures := map[string]*RuleTestFixture{}
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
				f, err := ruleTestFixtureFromFileInfo(fsys, inputsDir, e)
				if err != nil {
					return nil, err
				}
				fixtures[f.name] = f
			}
		}
	}
	t := &ruleTestDir{
		Dir:      ExistingDir(path),
		fixtures: fixtures,
	}
	return t, nil
}
