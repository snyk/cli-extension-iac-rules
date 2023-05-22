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
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-iac-rules/internal/project"
)

type (
	RelationFields struct {
		Name                  string
		PrimaryResourceType   string
		PrimaryAttributes     []string
		SecondaryResourceType string
		SecondaryAttributes   []string
	}

	RelationForm struct {
		Project *project.Project
		Fields  RelationFields
		Logger  *zerolog.Logger
	}
)

func (f *RelationForm) Run() error {
	if err := f.promptName(); err != nil {
		return err
	}
	if err := f.promptPrimaryResourceType(); err != nil {
		return err
	}
	if err := f.promptPrimaryAttributes(); err != nil {
		return err
	}
	if err := f.promptSecondaryResourceType(); err != nil {
		return err
	}
	if err := f.promptSecondaryAttributes(); err != nil {
		return err
	}

	relation, err := templateRelation(relationParams{
		Name:              f.Fields.Name,
		LeftResourceType:  f.Fields.PrimaryResourceType,
		LeftAttributes:    f.Fields.PrimaryAttributes,
		RightResourceType: f.Fields.SecondaryResourceType,
		RightAttributes:   f.Fields.SecondaryAttributes,
	})
	if err != nil {
		return err
	}
	path, err := f.Project.AddRelation(relation)
	if err != nil {
		return err
	}
	f.Logger.Info().Msgf("Adding relation '%s' to %s", f.Fields.Name, path)
	return nil
}

func (f *RelationForm) promptName() error {
	if f.Fields.Name != "" {
		return nil
	}

	prompt := textinput.New("Relation name:")
	name, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.Name = name
	return nil
}

func (f *RelationForm) promptPrimaryResourceType() error {
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

func (f *RelationForm) promptPrimaryAttributes() error {
	if len(f.Fields.PrimaryAttributes) > 0 {
		return nil
	}

	prompt := attrsPrompt(f.Fields.PrimaryResourceType)
	attrs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.PrimaryAttributes = attrs
	return nil
}

func (f *RelationForm) promptSecondaryResourceType() error {
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

func (f *RelationForm) promptSecondaryAttributes() error {
	if len(f.Fields.SecondaryAttributes) > 0 {
		return nil
	}

	prompt := attrsPrompt(f.Fields.SecondaryResourceType)
	attrs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.SecondaryAttributes = attrs
	return nil
}

func attrsPrompt(resourceType string) *multiplePrompt {
	return &multiplePrompt{
		prompt:  textinput.New(fmt.Sprintf("Attribute from %s:", resourceType)),
		another: confirmation.New("Would you like to add another attribute?", confirmation.No),
	}
}
