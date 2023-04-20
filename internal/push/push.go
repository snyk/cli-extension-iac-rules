package push

import (
	"context"
	"fmt"
	"os"

	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/afero"

	"github.com/snyk/cli-extension-cloud/internal/project"
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
	defaultOrgID := config.GetString(configuration.ORGANIZATION)

	prj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}

	manifest := prj.Manifest()
	fmt.Fprintf(os.Stderr, "Push config: %v\n", manifest.Push)
	if len(manifest.Push) == 0 {
		manifest.Push = []project.ManifestPush{
			{
				CustomRulesID:  "custom-rules-id",
				OrganizationID: defaultOrgID,
			},
		}
		prj.UpdateManifest(manifest)
		if err := prj.WriteChanges(); err != nil {
			return nil, err
		}
	}

	client := service.NewClient(
		ictx.GetNetworkAccess().GetHttpClient(),
		config.GetString(configuration.API_URL),
	)

	if err := client.CustomRules(ctx, defaultOrgID); err != nil {
		return nil, err
	}

	logger.Println("Hello world")
	fmt.Fprintf(os.Stderr, "Hello, world!\n")

	return []workflow.Data{}, nil
}
