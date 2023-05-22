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

package init

import (
	"fmt"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"
)

type TypeChoice string

const (
	TypeProject  = "project"
	TypeRule     = "rule"
	TypeSpec     = "rule spec"
	TypeRelation = "relation"
)

func TypeChoices() []TypeChoice {
	return []TypeChoice{
		TypeProject,
		TypeRule,
		TypeSpec,
		TypeRelation,
	}
}

func RegisterWorkflows(e workflow.Engine) error {
	workflowID := workflow.NewWorkflowIdentifier("iac.rules.init")
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-rules-init", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(workflowID, c, initWorkflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", workflowID, err)
	}
	return nil
}

func initWorkflow(
	ictx workflow.InvocationContext,
	input []workflow.Data,
) ([]workflow.Data, error) {
	prompt := selection.New("What do you want to initialize?", TypeChoices())
	choice, err := prompt.RunPrompt()
	if err != nil {
		return nil, err
	}
	switch choice {
	case TypeProject:
		return ProjectWorkflow(ictx, input)
	case TypeRule:
		return RuleWorkflow(ictx, input)
	case TypeSpec:
		return SpecWorkflow(ictx, input)
	case TypeRelation:
		return RelationWorkflow(ictx, input)
	default: // Should not happen
		return nil, fmt.Errorf("nothing to initialize")
	}
}
