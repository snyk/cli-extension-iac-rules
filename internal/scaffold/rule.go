package scaffold

import (
	"fmt"

	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/cli-extension-cloud/internal/scaffold/forms"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

var RuleWorkflowID = workflow.NewWorkflowIdentifier("iac.scaffold.rule")

func RuleWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	proj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}
	form := &forms.RuleForm{
		Project: proj,
	}
	if err := form.Run(); err != nil {
		return nil, err
	}
	if err := proj.WriteChanges(); err != nil {
		return nil, err
	}
	return []workflow.Data{}, nil
}

func RegisterRuleWorkflow(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-scaffold-rule", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(RuleWorkflowID, c, RuleWorkflow); err != nil {
		return fmt.Errorf("error while registering 'iac scaffold rule' workflow: %w", err)
	}
	return nil
}
