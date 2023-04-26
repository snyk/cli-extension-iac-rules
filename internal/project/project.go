package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/open-policy-agent/opa/ast"
	"github.com/snyk/policy-engine/pkg/data"
	"github.com/snyk/policy-engine/pkg/engine"
	"github.com/snyk/policy-engine/pkg/rego"
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
	return p.manifestFile.manifest
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
	ruleDirName, err := SafePackageName(ruleID)
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
	ruleDirName, err := SafePackageName(ruleID)
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

// AddRelation adds the given relation rule to the relations library for this
// project.
func (p *Project) AddRelation(contents string) error {
	return p.libDir.addRelation(contents)
}

// RelationNames returns the names of all relations defined in the project.
func (p *Project) RelationNames() ([]string, error) {
	ctx := context.Background()
	eng, err := p.engine(ctx)
	if err != nil {
		return nil, err
	}

	var relations []string
	err = eng.Query(ctx, &engine.QueryOptions{
		Query: "data.relations[_][_].name",
		ResultProcessor: func(v ast.Value) error {
			var relation string
			if err := rego.Bind(v, &relation); err != nil {
				return err
			}
			relations = append(relations, relation)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// RuleMetadata returns a map of rule ID to rule metadata from all rules in the
// project.
func (p *Project) RuleMetadata() (map[string]RuleMetadata, error) {
	ctx := context.Background()
	eng, err := p.engine(ctx)
	if err != nil {
		return nil, err
	}

	metadata := map[string]RuleMetadata{}
	for _, r := range eng.Metadata(ctx) {
		if r.Error != "" {
			continue
		}
		metadata[r.Metadata.ID] = RuleMetadata{
			ID:           r.Metadata.ID,
			Severity:     r.Metadata.Severity,
			Title:        r.Metadata.Title,
			Description:  r.Metadata.Description,
			Product:      r.Metadata.Product,
			Category:     r.Metadata.Category,
			Labels:       r.Metadata.Labels,
			Platform:     r.Metadata.Platform,
			ServiceGroup: r.Metadata.ServiceGroup,
		}
	}
	return metadata, nil
}

// InputTypeForRule returns the input for the given rule ID
func (p *Project) InputTypeForRule(ruleID string) (string, error) {
	ctx := context.Background()
	eng, err := p.engine(ctx)
	if err != nil {
		return "", err
	}

	var pkg string
	for _, r := range eng.Metadata(ctx) {
		if r.Error != "" {
			continue
		}
		pkg = r.Package
	}
	if pkg == "" {
		return "", err
	}

	var inputType string
	err = eng.Query(ctx, &engine.QueryOptions{
		Query: fmt.Sprintf("data.%s.input_type", pkg),
		ResultProcessor: func(v ast.Value) error {
			if err := rego.Bind(v, &inputType); err != nil {
				return err
			}
			return nil
		},
	})
	if err != nil {
		return "", err
	}
	return inputType, nil
}

func (p *Project) engine(ctx context.Context) (*engine.Engine, error) {
	fsys := afero.NewIOFS(p.FS)
	var providers []data.Provider
	if p.libDir.Exists() {
		providers = append(providers, data.FSProvider(fsys, p.libDir.Path()))
	}
	if p.rulesDir.Exists() {
		providers = append(providers, data.FSProvider(fsys, p.rulesDir.Path()))
	}
	eng := engine.NewEngine(ctx, &engine.EngineOptions{
		Providers: providers,
	})
	if len(eng.InitializationErrors) > 0 {
		return nil, &multierror.Error{
			Errors: eng.InitializationErrors,
		}
	}
	return eng, nil
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
