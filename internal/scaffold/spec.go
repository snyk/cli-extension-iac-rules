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

package scaffold

import (
	"fmt"

	"github.com/snyk/cli-extension-iac-rules/internal/project"
	"github.com/snyk/cli-extension-iac-rules/internal/scaffold/forms"
	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/snyk/policy-engine/pkg/input/cloudapi"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

var SpecWorkflowID = workflow.NewWorkflowIdentifier("iac.scaffold.spec")

func SpecWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	logger := ictx.GetEnhancedLogger()
	proj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}
	checkProject(proj, logger)
	config := ictx.GetConfiguration()
	client, err := cloudapi.NewClient(cloudapi.ClientConfig{
		HTTPClient: ictx.GetNetworkAccess().GetHttpClient(),
		URL:        config.GetString(configuration.API_URL),
	})
	if err != nil {
		return nil, err
	}
	form := &forms.SpecForm{
		Project: proj,
		Client:  client,
		OrgID:   config.GetString(configuration.ORGANIZATION),
		Logger:  logger,
	}
	if err := form.Run(); err != nil {
		return nil, err
	}
	if err := proj.WriteChanges(); err != nil {
		return nil, err
	}
	return []workflow.Data{}, nil
}

func RegisterSpecWorkflow(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-scaffold-spec", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(SpecWorkflowID, c, SpecWorkflow); err != nil {
		return fmt.Errorf("error while registering 'iac scaffold spec' workflow: %w", err)
	}
	return nil
}
