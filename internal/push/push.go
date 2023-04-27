package push

import (
	"bytes"
	"context"
	"fmt"

	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/snyk/policy-engine/pkg/bundle"
	"github.com/spf13/afero"

	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/cli-extension-cloud/internal/service"
)

var (
	WorkflowID = workflow.NewWorkflowIdentifier("iac.push")
)

const (
	FlagDelete = "delete"
)

func Workflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	ctx := context.Background()
	logger := ictx.GetLogger()
	config := ictx.GetConfiguration()
	currentOrgID := config.GetString(configuration.ORGANIZATION)
	del := ictx.GetConfiguration().GetBool(FlagDelete)

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

	client := service.NewClient(
		ictx.GetNetworkAccess().GetHttpClient(),
		config.GetString(configuration.API_URL),
	)
	manifest := prj.Manifest()
	push := GetManifestPushByOrganization(manifest, currentOrgID)
	if push == nil && del {
		return nil, fmt.Errorf("no rule bundle to delete")
	} else if push == nil {
		logger.Println("uploading new custom rules bundle")
		customRulesID, err := client.CreateCustomRules(ctx, currentOrgID, targz.Bytes())
		if err != nil {
			return nil, err
		}

		manifest.Push = append(manifest.Push, project.ManifestPush{
			CustomRulesID:  customRulesID,
			OrganizationID: currentOrgID,
		})
		prj.UpdateManifest(manifest)
		if err := prj.WriteChanges(); err != nil {
			return nil, err
		}
	} else if del {
		logger.Println("deleting custom rules bundle", push.CustomRulesID)
		err := client.DeleteCustomRules(ctx, push.OrganizationID, push.CustomRulesID)
		if err != nil {
			return nil, err
		}

		filtered := []project.ManifestPush{}
		for _, p := range manifest.Push {
			if p.OrganizationID != currentOrgID {
				filtered = append(filtered, p)
			}
		}
		manifest.Push = filtered
		prj.UpdateManifest(manifest)
		if err := prj.WriteChanges(); err != nil {
			return nil, err
		}
	} else {
		logger.Println("updating existing custom rules bundle", push.CustomRulesID)
		err := client.UpdateCustomRules(ctx, push.OrganizationID, push.CustomRulesID, targz.Bytes())
		if err != nil {
			return nil, err
		}
	}

	return []workflow.Data{}, nil
}

func GetManifestPushByOrganization(manifest project.Manifest, organizationID string) *project.ManifestPush {
	for _, push := range manifest.Push {
		if push.OrganizationID == organizationID {
			return &push
		}
	}
	return nil
}
