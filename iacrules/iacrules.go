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

package iacrules

import (
	"github.com/snyk/go-application-framework/pkg/local_workflows/config_utils"
	"github.com/snyk/go-application-framework/pkg/workflow"

	"github.com/snyk/cli-extension-iac-rules/internal/constants"
	initWorkflow "github.com/snyk/cli-extension-iac-rules/internal/init"
	"github.com/snyk/cli-extension-iac-rules/internal/push"
	"github.com/snyk/cli-extension-iac-rules/internal/repl"
	"github.com/snyk/cli-extension-iac-rules/internal/test"
)

func Init(e workflow.Engine) error {
	if err := initWorkflow.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := test.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := push.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := repl.RegisterWorkflows(e); err != nil {
		return err
	}
	config_utils.AddFeatureFlagToConfig(e, constants.FF_IAC_NEW_ENGINE, constants.FF_IAC_NEW_ENGINE)
	return nil
}
