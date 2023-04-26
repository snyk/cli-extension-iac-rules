package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/cli-extension-cloud/internal/scaffold/forms"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

var ProjectWorkflowID = workflow.NewWorkflowIdentifier("iac.scaffold")

func ProjectWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	proj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}
	defaultName := filepath.Base(wd)
	if name := proj.Manifest().Name; name != "" {
		defaultName = name
	}
	form := &forms.ProjectForm{
		Project:     proj,
		DefaultName: defaultName,
	}
	if err := form.Run(); err != nil {
		return nil, err
	}
	if err := proj.WriteChanges(); err != nil {
		return nil, err
	}
	return []workflow.Data{}, nil
}

func RegisterProjectWorkflow(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-scaffold", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(ProjectWorkflowID, c, ProjectWorkflow); err != nil {
		return fmt.Errorf("error while registering 'iac scaffold' workflow: %w", err)
	}
	return nil
}
