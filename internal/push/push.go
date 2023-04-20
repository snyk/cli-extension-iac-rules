package push

import (
	"context"
	"fmt"
	"os"

	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"

	"github.com/snyk/cli-extension-cloud/internal/service"
)

var (
	WorkflowID = workflow.NewWorkflowIdentifier("cloud.rules.push")
)

func Workflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	ctx := context.Background()
	logger := ictx.GetLogger()
	config := ictx.GetConfiguration()
	orgID := config.GetString(configuration.ORGANIZATION)
	client := service.NewClient(
		ictx.GetNetworkAccess().GetHttpClient(),
		config.GetString(configuration.API_URL),
	)

	if err := client.CustomRules(ctx, orgID); err != nil {
		return nil, err
	}

	logger.Println("Hello world")
	fmt.Fprintf(os.Stderr, "Hello, world!\n")

	return []workflow.Data{}, nil
}
