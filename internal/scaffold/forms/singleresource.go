package forms

import (
	"encoding/json"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/snyk/cli-extension-cloud/internal/project"
)

type (
	SingleResourceRuleFields struct {
		ResourceType string
	}

	SingleResourceRuleForm struct {
		Project   *project.Project
		RuleID    string
		InputType string
		Metadata  *project.RuleMetadata
		Fields    SingleResourceRuleFields
	}
)

func (f *SingleResourceRuleForm) Run() error {
	if err := f.promptResourceType(); err != nil {
		return err
	}

	metadataJSON, err := json.MarshalIndent(f.Metadata, "", "\t")
	if err != nil {
		return err
	}
	rulePackage, err := project.SafePackageName(f.RuleID)
	if err != nil {
		return err
	}
	rule, err := templateSingleResourceRule(singleResourceRuleParams{
		RulePackage:  rulePackage,
		InputType:    f.InputType,
		RuleMetadata: string(metadataJSON),
		ResourceType: f.Fields.ResourceType,
	})
	if err != nil {
		return err
	}
	return f.Project.AddRule(f.RuleID, "main.rego", rule)
}

func (f *SingleResourceRuleForm) promptResourceType() error {
	if f.Fields.ResourceType != "" {
		return nil
	}

	prompt := textinput.New("Resource type")
	resourceType, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.ResourceType = resourceType
	return nil
}
