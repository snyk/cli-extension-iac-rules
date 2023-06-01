// © 2023 Snyk Limited All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/spf13/pflag"

	"github.com/snyk/cli-extension-iac-rules/internal/project"
	"github.com/snyk/cli-extension-iac-rules/internal/service"
)

var ()

const (
	flagDelete = "delete"
)

func RegisterWorkflows(e workflow.Engine) error {
	workflowID := workflow.NewWorkflowIdentifier("iac.rules.push")
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-rules-push", pflag.ExitOnError)

	flagset.Bool(flagDelete, false, "Delete upstream rule bundle")

	c := workflow.ConfigurationOptionsFromFlagset(flagset)

	if _, err := e.Register(workflowID, c, pushWorkflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", workflowID, err)
	}
	return nil
}

func pushWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	ctx := context.Background()
	logger := ictx.GetLogger()
	config := ictx.GetConfiguration()
	currentOrgID := config.GetString(configuration.ORGANIZATION)
	del := ictx.GetConfiguration().GetBool(flagDelete)

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
	push := getManifestPushByOrganization(manifest, currentOrgID)
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

	fmt.Fprintln(os.Stderr, "Successfully uploaded custom rule bundle.")
	return []workflow.Data{}, nil
}

func getManifestPushByOrganization(manifest project.Manifest, organizationID string) *project.ManifestPush {
	for _, push := range manifest.Push {
		if push.OrganizationID == organizationID {
			return &push
		}
	}
	return nil
}
