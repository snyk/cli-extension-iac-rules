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

	"github.com/erikgeiser/promptkit/selection"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/pflag"
)

type ScaffoldChoice string

const (
	ScaffoldProject  = "project"
	ScaffoldRule     = "rule"
	ScaffoldSpec     = "rule spec"
	ScaffoldRelation = "relation"
)

func ScaffoldChoices() []ScaffoldChoice {
	return []ScaffoldChoice{
		ScaffoldProject,
		ScaffoldRule,
		ScaffoldSpec,
		ScaffoldRelation,
	}
}

var ScaffoldWorkflowID = workflow.NewWorkflowIdentifier("iac.scaffold")

func RegisterWorkflows(e workflow.Engine) error {
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-scaffold", pflag.ExitOnError)
	c := workflow.ConfigurationOptionsFromFlagset(flagset)
	if _, err := e.Register(ScaffoldWorkflowID, c, ScaffoldWorkflow); err != nil {
		return fmt.Errorf("error while registering 'iac scaffold' workflow: %w", err)
	}
	return nil
}

func ScaffoldWorkflow(
	ictx workflow.InvocationContext,
	input []workflow.Data,
) ([]workflow.Data, error) {
	prompt := selection.New("What do you want to scaffold?", ScaffoldChoices())
	choice, err := prompt.RunPrompt()
	if err != nil {
		return nil, err
	}
	switch choice {
	case ScaffoldProject:
		return ProjectWorkflow(ictx, input)
	case ScaffoldRule:
		return RuleWorkflow(ictx, input)
	case ScaffoldSpec:
		return SpecWorkflow(ictx, input)
	case ScaffoldRelation:
		return RelationWorkflow(ictx, input)
	default: // Should not happen
		return nil, fmt.Errorf("nothing to scaffold")
	}
}
