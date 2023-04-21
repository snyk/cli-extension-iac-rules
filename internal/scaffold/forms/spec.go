package forms

import (
	"sort"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/input/cloudapi"
)

type (
	SpecFields struct {
		RuleID    string
		Name      string
		InputType string
	}

	SpecForm struct {
		Project *project.Project
		Client  *cloudapi.Client
		OrgID   string
		Fields  SpecFields
	}
)

func (f *SpecForm) Run() error {
	if err := f.promptRuleID(); err != nil {
		return err
	}
	if err := f.promptName(); err != nil {
		return err
	}
	if err := f.promptInputType(); err != nil {
		return err
	}

	if f.Fields.InputType == input.CloudScan.Name {
		form := &CloudSpecForm{
			Project: f.Project,
			Client:  f.Client,
			OrgID:   f.OrgID,
			RuleID:  f.Fields.RuleID,
			Name:    f.Fields.Name,
		}
		return form.Run()
	} else {
		filename, contents := specForInputType(f.Fields.InputType, f.Fields.Name)
		return f.Project.AddRuleTestFixture(f.Fields.RuleID, filename, contents)
	}
}

func (f *SpecForm) promptRuleID() error {
	if f.Fields.RuleID != "" {
		return nil
	}

	const enterManually = "Enter manually"
	metadata := f.Project.Metadata()
	var choices []string
	for id := range metadata {
		choices = append(choices, id)
	}
	sort.Strings(choices)
	choices = append(choices, enterManually)
	prompt := selection.New("Choose a rule ID", choices)
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	switch choice {
	case enterManually:
		prompt := textinput.New("Rule ID")
		ruleID, err := prompt.RunPrompt()
		if err != nil {
			return err
		}
		f.Fields.RuleID = ruleID
	default:
		f.Fields.RuleID = choice
	}
	return nil
}

func (f *SpecForm) promptName() error {
	if f.Fields.Name != "" {
		return nil
	}

	prompt := textinput.New("Spec name")
	name, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.Name = name
	return nil
}

func (f *SpecForm) promptInputType() error {
	if f.Fields.InputType != "" {
		return nil
	}

	var choices []string
	defaultInputType := f.Project.InputTypeForRule(f.Fields.RuleID)
	if defaultInputType != "" {
		choices = []string{defaultInputType}
		for _, t := range inputTypes() {
			if t != defaultInputType {
				choices = append(choices, t)
			}
		}
	} else {
		choices = inputTypes()
	}
	prompt := selection.New("Input type", choices)
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.InputType = choice
	return nil
}
