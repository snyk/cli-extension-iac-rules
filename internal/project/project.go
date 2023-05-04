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
	specDir      *specDir
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
	if err := p.specDir.WriteChanges(p.FS); err != nil {
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
func (p *Project) AddRule(ruleID string, regoFileName string, contents []byte) (string, error) {
	ruleDirName, err := SafePackageName(ruleID)
	if err != nil {
		return "", err
	}
	safeRegoFileName, err := safeFilename(regoFileName)
	if err != nil {
		return "", err
	}
	return p.rulesDir.addRule(ruleDirName, safeRegoFileName, contents)
}

// AddRuleSpec adds a rule to the project. The given rule ID will be transformed
// to a valid package name and the spec name will be transformed to fit similar
// constraints.
func (p *Project) AddRuleSpec(ruleID string, name string, contents []byte) (string, error) {
	ruleDirName, err := SafePackageName(ruleID)
	if err != nil {
		return "", err
	}
	safeName, err := safeFilename(name)
	if err != nil {
		return "", err
	}
	return p.specDir.addRuleSpec(ruleDirName, safeName, contents)
}

// RuleSpecs returns the rule specs in the project. The returned fixtures can be
// modified in-place, then the changes can be persisted by calling WriteChanges
// on the project.
func (p *Project) RuleSpecs() []*RuleSpec {
	return p.specDir.fixtures()
}

// AddRelation adds the given relation rule to the relations library for this
// project.
func (p *Project) AddRelation(contents string) (string, error) {
	return p.libDir.addRelation(contents)
}

// RelationNames returns the names of all relations defined in the project.
func (p *Project) RelationNames() ([]string, error) {
	ctx := context.Background()
	eng, err := p.Engine(ctx)
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
	eng, err := p.Engine(ctx)
	if err != nil {
		return nil, err
	}

	metadata := map[string]RuleMetadata{}
	for _, r := range eng.Metadata(ctx) {
		if r.Error != "" {
			return nil, fmt.Errorf(r.Error)
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
	eng, err := p.Engine(ctx)
	if err != nil {
		return "", err
	}

	var pkg string
	for _, r := range eng.Metadata(ctx) {
		if r.Error != "" {
			continue
		}
		if r.Metadata.ID != ruleID {
			continue
		}
		pkg = r.Package
	}
	if pkg == "" {
		return "", err
	}

	var inputType string
	err = eng.Query(ctx, &engine.QueryOptions{
		Query: fmt.Sprintf("%s.input_type", pkg),
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

func (p *Project) Engine(ctx context.Context) (*engine.Engine, error) {
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
	spec, err := specFromDir(fsys, root)
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
		specDir:      spec,
		manifestFile: manifest,
	}
	return p, nil
}
