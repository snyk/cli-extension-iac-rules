package cloud

import (
	"fmt"

	"github.com/snyk/cli-extension-cloud/internal/scaffold"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"

	"github.com/snyk/cli-extension-cloud/internal/push"
)

func Init(e workflow.Engine) error {
	if err := scaffold.RegisterWorkflows(e); err != nil {
		return err
	}

	if err := InitPush(e); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", push.WorkflowID, err)
	}

	return nil
}

func InitPush(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-cloud-push", pflag.ExitOnError)

	flagset.Bool(push.FlagDelete, false, "Delete upstream rule bundle")

	c := workflow.ConfigurationOptionsFromFlagset(flagset)

	if _, err := e.Register(push.WorkflowID, c, push.Workflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", push.WorkflowID, err)
	}

	return nil
}
