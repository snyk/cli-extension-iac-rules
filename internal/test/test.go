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

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/snyk/go-application-framework/pkg/configuration"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/snyk/policy-engine/pkg/engine"
	"github.com/snyk/policy-engine/pkg/models"
	"github.com/snyk/policy-engine/pkg/postprocess"
	"github.com/snyk/policy-engine/pkg/rego/test"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/snyk/cli-extension-iac-rules/internal/project"
	"github.com/snyk/cli-extension-iac-rules/internal/utils"
)

const (
	flagUpdateExpected = "update-expected"
)

func RegisterWorkflows(e workflow.Engine) error {
	workflowID := workflow.NewWorkflowIdentifier("iac.rules.test")
	flagset := pflag.NewFlagSet("snyk-cli-extension-iac-rules-test", pflag.ExitOnError)

	flagset.Bool(flagUpdateExpected, false, "Updated expected JSON files based on actual results")

	c := workflow.ConfigurationOptionsFromFlagset(flagset)

	if _, err := e.Register(workflowID, c, testWorkflow); err != nil {
		return fmt.Errorf("error while registering %s workflow: %w", workflowID, err)
	}
	return nil
}

func testWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	ctx := context.Background()
	logger := ictx.GetLogger()
	verbose := ictx.GetConfiguration().GetBool(configuration.DEBUG)

	updateExpected := ictx.GetConfiguration().GetBool(flagUpdateExpected)
	fixturesFailed := 0
	fixturesTested := 0

	fs := afero.NewOsFs()
	prj, err := project.FromDir(fs, ".")
	if err != nil {
		return nil, err
	}

	eng, err := prj.Engine(ctx)
	if err != nil {
		return nil, err
	}

	ruleDirNameToRuleID, err := makeRuleDirNameToRuleID(eng, ctx)
	if err != nil {
		return nil, err
	}

	for _, fixture := range prj.RuleSpecs() {
		ruleID, ok := ruleDirNameToRuleID[fixture.RuleDirName]
		if !ok {
			return nil, fmt.Errorf("ID metadata not found for %s", fixture.RuleDirName)
		}

		actualResults, err := runEngine(eng, ruleID, fixture.Input.Path())
		if err != nil {
			return nil, fmt.Errorf("Error running engine on %v: %w", fixture.Input.Path(), err)
		}
		actualBytes, err := json.MarshalIndent(actualResults, "", "  ")
		actual := string(actualBytes)

		var expected string
		expectedPath := fixture.ExpectedPath()
		expectedFile, err := fs.Open(expectedPath)
		if err == nil {
			expectedBytes, err := io.ReadAll(expectedFile)
			if err != nil {
				return nil, err
			}
			expected = string(expectedBytes)
			expectedFile.Close()
		}

		if expected != actual {
			fixturesFailed += 1
			edits := myers.ComputeEdits(span.URI(expectedPath), expected, actual)
			diff := gotextdiff.ToUnified(expectedPath, fixture.Input.Path(), expected, edits)
			fmt.Fprintf(os.Stderr, "expected output does not match for rule %s\n: %s", ruleID, diff)

			if updateExpected {
				if err := os.MkdirAll(filepath.Dir(expectedPath), 0755); err != nil {
					return nil, err
				}
				fixture.UpdateExpected(actualBytes)
				if err := fixture.WriteChanges(fs); err != nil {
					return nil, err
				}
			}
		}

		fixturesTested += 1
	}

	logger.Println(fixturesFailed, "spec files failed")
	logger.Println(fixturesTested, "spec files tested")

	result, err := test.Test(ctx, test.Options{
		Providers: prj.Providers(),
		Verbose:   verbose,
	})
	if err != nil {
		return nil, err
	}
	if !result.Passed {
		return nil, fmt.Errorf("tests failed")
	}

	return []workflow.Data{}, nil
}

func makeRuleDirNameToRuleID(eng *engine.Engine, ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	for _, mdr := range eng.Metadata(ctx) {
		if ruleID := mdr.Metadata.ID; ruleID != "" {
			ruleDirName, err := project.RuleIDToSafeFileName(ruleID)
			if err != nil {
				return nil, err
			}
			out[ruleDirName] = ruleID
		}
	}
	return out, nil
}

func runEngine(eng *engine.Engine, ruleID string, path string) ([]models.RuleResult, error) {
	singleInput, err := utils.LoadSingleInput(path)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	inputs := []models.State{singleInput.State}
	results := eng.Eval(ctx, &engine.EvalOptions{
		Inputs:  inputs,
		RuleIDs: []string{ruleID},
	})
	postprocess.AddSourceLocs(results, singleInput.Loader)

	if len(results.Results) != 1 || len(results.Results[0].RuleResults) != 1 {
		return nil, fmt.Errorf("internal error: expected a single rule result")
	}
	return results.Results[0].RuleResults[0].Results, nil
}
