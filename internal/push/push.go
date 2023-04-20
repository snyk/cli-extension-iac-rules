package push

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/snyk/policy-engine/pkg/bundle"
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
	bundled, err := bundle.BuildBundle(bundle.NewDirReader(prj.Path()))
	if err != nil {
		return nil, err
	}
	if err := bundled.Validate(); err != nil {
		return nil, err
	}
	logger.Println("validated bundle")

	targz := &bytes.Buffer{}
	if err := bundle.NewTarGzWriter(targz).Write(bundled); err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Bundle bytes: %d\n", len(targz.Bytes()))

	client := service.NewClient(
		ictx.GetNetworkAccess().GetHttpClient(),
		config.GetString(configuration.API_URL),
	)
	manifest := prj.Manifest()
	if len(manifest.Push) == 0 {
		logger.Println("uploading new custom rules bundle")
		customRulesID, err := client.CreateCustomRules(ctx, defaultOrgID, targz.Bytes())
		if err != nil {
			return nil, err
		}

		manifest.Push = []project.ManifestPush{
			{
				CustomRulesID:  customRulesID,
				OrganizationID: defaultOrgID,
			},
		}
		prj.UpdateManifest(manifest)
		if err := prj.WriteChanges(); err != nil {
			return nil, err
		}
	}

	return []workflow.Data{}, nil
}
