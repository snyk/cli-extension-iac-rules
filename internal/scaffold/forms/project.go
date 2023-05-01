package forms

import (
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-cloud/internal/project"
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
