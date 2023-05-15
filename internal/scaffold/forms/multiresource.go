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

package forms

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/open-policy-agent/opa/ast"
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-iac-rules/internal/project"
)

type (
	MultiResourceRuleFields struct {
		PrimaryResourceType   string
		SecondaryResourceType string
		Relation              string
	}

	MultiResourceRuleForm struct {
		Project   *project.Project
		RuleID    string
		InputType string
		Metadata  *project.RuleMetadata
		Fields    MultiResourceRuleFields
		Logger    *zerolog.Logger
	}
)

func (f *MultiResourceRuleForm) Run() error {
	if err := f.promptPrimaryResourceType(); err != nil {
		return err
	}
	if err := f.promptSecondaryResourceType(); err != nil {
		return err
	}
	if err := f.promptRelation(); err != nil {
		return err
	}

	pluralizer := pluralize.NewClient()
	primarySingular, primaryPlural := toSingularAndPlural(pluralizer, f.InputType, f.Fields.PrimaryResourceType)
	if isReserved(primarySingular) || isReserved(primaryPlural) {
		// Sensible fallbacks for illegal or conflicting variable names
		primaryPlural = "primary_resources"
		primarySingular = "primary"
	}
	secondarySingular, secondaryPlural := toSingularAndPlural(pluralizer, f.InputType, f.Fields.SecondaryResourceType)
	if isReserved(secondarySingular) ||
		isReserved(secondaryPlural) ||
		secondarySingular == primarySingular ||
		secondaryPlural == primaryPlural {
		secondarySingular = "secondary_resources"
		secondaryPlural = "secondary"
	}
	metadataJSON, err := json.MarshalIndent(f.Metadata, "", "\t")
	if err != nil {
		return err
	}
	rulePackage, err := project.SafePackageName(f.RuleID)
	if err != nil {
		return err
	}
	rule, err := templateMultiResourceRule(multiResourceRuleParams{
		RulePackage:               rulePackage,
		InputType:                 f.InputType,
		RuleMetadata:              string(metadataJSON),
		PrimaryResourceType:       f.Fields.PrimaryResourceType,
		PrimaryResourceSingular:   primarySingular,
		PrimaryResourcePlural:     primaryPlural,
		SecondaryResourceSingular: secondarySingular,
		SecondaryResourcePlural:   secondaryPlural,
		Relation:                  f.Fields.Relation,
	})
	if err != nil {
		return err
	}
	path, err := f.Project.AddRule(f.RuleID, "main.rego", rule)
	if err != nil {
		return err
	}
	f.Logger.Info().Msgf("Writing rule to %s", path)
	return nil
}

func (f *MultiResourceRuleForm) promptPrimaryResourceType() error {
	if f.Fields.PrimaryResourceType != "" {
		return nil
	}

	prompt := textinput.New("Primary resource type:")
	primary, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.PrimaryResourceType = primary
	return nil
}

func (f *MultiResourceRuleForm) promptSecondaryResourceType() error {
	if f.Fields.SecondaryResourceType != "" {
		return nil
	}

	prompt := textinput.New("Secondary resource type:")
	secondary, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.SecondaryResourceType = secondary
	return nil
}

func (f *MultiResourceRuleForm) promptRelation() error {
	if f.Fields.Relation != "" {
		return nil
	}

	const addNewRelation = "Add a new relation"
	const enterManually = "Enter manually"
	choices := []string{addNewRelation, enterManually}
	relations, err := f.Project.RelationNames()
	if err != nil {
		// we don't want to crash if there's a compilation or execution error.
		// instead, we'll still allow users to add a new relation or enter in a
		// relation manually.
		relations = []string{}
	}
	prompt := selection.New("Choose a relation:", append(relations, choices...))
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	switch choice {
	case addNewRelation:
		form := &RelationForm{
			Project: f.Project,
			Logger:  f.Logger,
			Fields: RelationFields{
				PrimaryResourceType:   f.Fields.PrimaryResourceType,
				SecondaryResourceType: f.Fields.SecondaryResourceType,
			},
		}
		if err := form.Run(); err != nil {
			return err
		}
		f.Fields.Relation = form.Fields.Name
	case enterManually:
		prompt := textinput.New("Relation name:")
		relation, err := prompt.RunPrompt()
		if err != nil {
			return err
		}
		f.Fields.Relation = relation
	default:
		f.Fields.Relation = choice
	}
	return nil
}

var camelPat = regexp.MustCompile(`(.)([A-Z])`)

func toSnakeCase(s string) string {
	out := camelPat.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(out)
}

// This is an aesthetic choice. Some words that end in "ing" can look strange in
// the plural form, e.g. "loggings". Appending "_config" looks nicer.
func fixGerund(s string) string {
	if strings.HasSuffix(s, "ing") {
		return fmt.Sprintf("%s_config", s)
	}
	return s
}

func toSingularAndPlural(client *pluralize.Client, inputType, resourceType string) (string, string) {
	switch inputType {
	case "k8s":
		resourceType = toSnakeCase(resourceType)
	case "cfn":
		split := strings.Split(resourceType, "::")
		resourceType = toSnakeCase(split[len(split)-1])
	case "arm":
		split := strings.Split(resourceType, "/")
		resourceType = toSnakeCase(split[len(split)-1])
	}
	split := strings.Split(resourceType, "_")
	suffix := split[len(split)-1]
	suffix = fixGerund(suffix)
	return client.Singular(suffix), client.Plural(suffix)
}

func isReserved(s string) bool {
	return ast.IsKeyword(s) ||
		ast.ReservedVars.Contains(ast.Var(s)) ||
		s == "resources" ||
		s == "deny" ||
		s == "snyk" ||
		s == ""
}
