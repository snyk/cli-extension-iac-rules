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
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-iac-rules/internal/project"
)

type (
	ProjectFields struct {
		Name string
	}

	ProjectForm struct {
		Project     *project.Project
		DefaultName string
		Fields      ProjectFields
		Logger      *zerolog.Logger
	}
)

func (p *ProjectForm) Run() error {
	if err := p.promptName(); err != nil {
		return err
	}

	manifest := p.Project.Manifest()
	manifest.Name = p.Fields.Name

	p.Project.UpdateManifest(manifest)
	p.Logger.Info().Msgf("Initializing or updating project '%s'", p.Fields.Name)
	return nil
}

func (p *ProjectForm) promptName() error {
	if p.Fields.Name != "" {
		return nil
	}

	prompt := textinput.New("Project name:")
	prompt.InitialValue = p.DefaultName
	name, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	p.Fields.Name = name
	return nil
}
