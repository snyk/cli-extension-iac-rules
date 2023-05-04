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

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-cloud/internal/project"
)

type (
	SingleResourceRuleFields struct {
		ResourceType string
	}

	SingleResourceRuleForm struct {
		Project   *project.Project
		RuleID    string
		InputType string
		Metadata  *project.RuleMetadata
		Fields    SingleResourceRuleFields
		Logger    *zerolog.Logger
	}
)

func (f *SingleResourceRuleForm) Run() error {
	if err := f.promptResourceType(); err != nil {
		return err
	}

	metadataJSON, err := json.MarshalIndent(f.Metadata, "", "\t")
	if err != nil {
		return err
	}
	rulePackage, err := project.SafePackageName(f.RuleID)
	if err != nil {
		return err
	}
	rule, err := templateSingleResourceRule(singleResourceRuleParams{
		RulePackage:  rulePackage,
		InputType:    f.InputType,
		RuleMetadata: string(metadataJSON),
		ResourceType: f.Fields.ResourceType,
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

func (f *SingleResourceRuleForm) promptResourceType() error {
	if f.Fields.ResourceType != "" {
		return nil
	}

	prompt := textinput.New("Resource type:")
	resourceType, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.ResourceType = resourceType
	return nil
}
