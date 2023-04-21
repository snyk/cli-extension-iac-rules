package scaffold

import (
	"fmt"

	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/cli-extension-cloud/internal/scaffold/forms"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

var RelationWorkflowID = workflow.NewWorkflowIdentifier("iac.scaffold.relation")

func RelationWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	proj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}
	form := &forms.RelationForm{Project: proj}
	if err := form.Run(); err != nil {
		return nil, err
	}
	if err := proj.WriteChanges(); err != nil {
		return nil, err
	}
	return []workflow.Data{}, nil
}

func RegisterRelationWorkflow(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-scaffold-relation", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(RelationWorkflowID, c, RelationWorkflow); err != nil {
		return fmt.Errorf("error while registering 'iac scaffold relation' workflow: %w", err)
	}
	return nil
}
