package cloud

import (
	"fmt"

	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"
)

var (
	WorkflowID = workflow.NewWorkflowIdentifier("cloud.foo")
)

func CloudWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) (sbomDocs []workflow.Data, err error) {
	logger := ictx.GetLogger()

	logger.Println("Hello world")

	return []workflow.Data{}, nil
}

func Init(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-cloud", pflag.ExitOnError)

	c := workflow.ConfigurationOptionsFromFlagset(flagset)

	if _, err := e.Register(WorkflowID, c, CloudWorkflow); err != nil {
		return fmt.Errorf("error while registering SBOM workflow: %w", err)
	}

	return nil
}
