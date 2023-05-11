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

package repl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/snyk/policy-engine/pkg/rego/repl"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/snyk/cli-extension-iac-rules/internal/project"
	"github.com/snyk/cli-extension-iac-rules/internal/utils"
)

const (
	flagInit  = "repl-init"
	flagInput = "repl-input"
)

func RegisterWorkflows(e workflow.Engine) error {
	workflowID := workflow.NewWorkflowIdentifier("iac.repl")

	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-repl", pflag.ExitOnError)
	flagset.StringSlice(flagInit, []string{}, "Run commands on REPL initialization")
	flagset.String(flagInput, "", "Input IaC file")

	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(workflowID, c, replWorkflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", workflowID, err)
	}
	return nil
}

func replWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	ctx := context.Background()
	init := ictx.GetConfiguration().GetStringSlice(flagInit)
	inputPath := ictx.GetConfiguration().GetString(flagInput)

	fs := afero.NewOsFs()
	prj, err := project.FromDir(fs, ".")
	if err != nil {
		return nil, err
	}

	input := map[string]interface{}{}
	if inputPath != "" {
		singleInput, err := utils.LoadSingleInput(inputPath)
		if err != nil {
			return nil, err
		}
		bytes, err := json.Marshal(singleInput.State)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, &input); err != nil {
			return nil, err
		}
	}

	err = repl.Repl(ctx, repl.Options{
		Providers: prj.Providers(),
		Init:      init,
		Input:     input,
	})
	if err != nil {
		return nil, err
	}

	return []workflow.Data{}, nil
}
