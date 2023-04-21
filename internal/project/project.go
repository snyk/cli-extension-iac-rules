package project

import (
	"github.com/spf13/afero"
)

// Project represents a custom rules project directory. It encapsulates all
// operations for creating and updating the contents of that directory.
type Project struct {
	*Dir
	FS           afero.Fs
	rulesDir     *rulesDir
	libDir       *libDir
	testsDir     *testsDir
	manifestFile *manifestFile
}

// WriteChanges persists any changes to this project back to disk. This
// operation is (essentially) idempotent.
func (p *Project) WriteChanges() error {
	if err := p.Dir.WriteChanges(p.FS); err != nil {
		return err
	}
	if err := p.rulesDir.WriteChanges(p.FS); err != nil {
		return err
	}
	if err := p.libDir.WriteChanges(p.FS); err != nil {
		return err
	}
	if err := p.testsDir.WriteChanges(p.FS); err != nil {
		return err
	}
	if err := p.manifestFile.WriteChanges(p.FS); err != nil {
		return err
	}
	return nil
}

// Manifest retrieves a copy of the project's manifest.
func (p *Project) Manifest() Manifest {
	return p.manifestFile.manifest.copy()
}

// UpdateManifest updates the project's manifest.
func (p *Project) UpdateManifest(m Manifest) {
	p.manifestFile.UpdateContents(m)
}

// ListRules lists the rule directories in the project.
func (p *Project) ListRules() []string {
	return p.rulesDir.ruleDirNames()
}

// AddRule adds a rule to the project. The given rule ID will be transformed to
// a valid package name and the rego filename will be transformed to fit
// similar constraints.
func (p *Project) AddRule(ruleID string, regoFileName string, contents []byte) error {
	ruleDirName, err := safePackageName(ruleID)
	if err != nil {
		return err
	}
	safeRegoFileName, err := safeFilename(regoFileName)
	if err != nil {
		return err
	}
	return p.rulesDir.addRule(ruleDirName, safeRegoFileName, contents)
}

// AddRuleTestFixture adds a rule to the project. The given rule ID will be
// transformed to a valid package name and the test fixture name will be
// transformed to fit similar constraints.
func (p *Project) AddRuleTestFixture(ruleID string, name string, contents []byte) error {
	ruleDirName, err := safePackageName(ruleID)
	if err != nil {
		return err
	}
	safeName, err := safeFilename(name)
	if err != nil {
		return err
	}
	return p.testsDir.addRuleTestFixture(ruleDirName, safeName, contents)
}

// RuleTestFixtures returns the test fixtures in the project. The returned
// fixtures can be modified in-place, then the changes can be persisted by
// calling WriteChanges on the project.
func (p *Project) RuleTestFixtures() []*RuleTestFixture {
	return p.testsDir.fixtures()
}

// FromDir returns a Project object from the given directory, whether it exists
// or not.
func FromDir(fsys afero.Fs, root string) (*Project, error) {
	dir, err := DirFromPath(fsys, root)
	if err != nil {
		return nil, err
	}
	rules, err := rulesFromDir(fsys, root)
	if err != nil {
		return nil, err
	}
	lib, err := libFromDir(fsys, root)
	if err != nil {
		return nil, err
	}
	tests, err := testsFromDir(fsys, root)
	if err != nil {
		return nil, err
	}
	manifest, err := manifestFromDir(fsys, root)
	if err != nil {
		return nil, err
	}
	p := &Project{
		Dir:          dir,
		FS:           fsys,
		rulesDir:     rules,
		libDir:       lib,
		testsDir:     tests,
		manifestFile: manifest,
	}
	return p, nil
}
