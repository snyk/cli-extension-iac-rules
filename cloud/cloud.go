package cloud

import (
	"fmt"
	"os"

	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"

	"github.com/snyk/cli-extension-cloud/internal/push"
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
	fmt.Fprintf(os.Stderr, "Hello, world!\n")

	return []workflow.Data{}, nil
}

func Init(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-cloud", pflag.ExitOnError)

	c := workflow.ConfigurationOptionsFromFlagset(flagset)

	if _, err := e.Register(WorkflowID, c, CloudWorkflow); err != nil {
		return fmt.Errorf("error while registering SBOM workflow: %w", err)
	}

	if _, err := e.Register(push.WorkflowID, c, push.Workflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", push.WorkflowID, err)
	}

	return nil
}
