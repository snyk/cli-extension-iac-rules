package forms

import (
	"sort"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rs/zerolog"
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
		Logger  *zerolog.Logger
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
			Logger:  f.Logger,
		}
		return form.Run()
	} else {
		filename, contents := specForInputType(f.Fields.InputType, f.Fields.Name)
		path, err := f.Project.AddRuleSpec(f.Fields.RuleID, filename, contents)
		if err != nil {
			return err
		}
		f.Logger.Info().Msgf("Writing rule spec stub to %s", path)
		return nil
	}
}

func (f *SpecForm) promptRuleID() error {
	if f.Fields.RuleID != "" {
		return nil
	}

	var choices []string
	metadata, err := f.Project.RuleMetadata()
	if err == nil {
		for id := range metadata {
			choices = append(choices, id)
		}
	}
	sort.Strings(choices)
	const enterManually = "Enter manually"
	choices = append(choices, enterManually)
	prompt := selection.New("Choose a rule ID:", choices)
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	switch choice {
	case enterManually:
		prompt := textinput.New("Rule ID:")
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

	prompt := textinput.New("Spec name:")
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
	defaultInputType, err := f.Project.InputTypeForRule(f.Fields.RuleID)
	if err == nil && defaultInputType != "" {
		choices = []string{defaultInputType}
		for _, t := range inputTypes() {
			if t != defaultInputType {
				choices = append(choices, t)
			}
		}
	} else {
		choices = inputTypes()
	}
	prompt := selection.New("Input type:", choices)
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.InputType = choice
	return nil
}
