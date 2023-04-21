package forms

import (
	"context"
	"encoding/json"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/snyk/cli-extension-cloud/internal/project"
	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/input/cloudapi"
)

type (
	CloudSpecFields struct {
		ResourceTypes  []string
		NativeIDs      []string
		EnvironmentIDs []string
		Locations      []string
	}

	CloudSpecForm struct {
		Project *project.Project
		Client  *cloudapi.Client
		OrgID   string
		RuleID  string
		Name    string
		Fields  CloudSpecFields
	}
)

func (f *CloudSpecForm) Run() error {
	if err := f.promptTopLevelFilter(); err != nil {
		return err
	}

	ctx := context.Background()
	loader := input.CloudLoader{
		Client: f.Client,
	}
	state, err := loader.GetState(ctx, f.OrgID, cloudapi.ResourcesParameters{
		EnvironmentID: f.Fields.EnvironmentIDs,
		ResourceType:  f.Fields.ResourceTypes,
		NativeID:      f.Fields.NativeIDs,
		Location:      f.Fields.Locations,
	})
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	filename := addExtIfNeeded(f.Name, ".json")
	f.Project.AddRuleTestFixture(f.RuleID, filename, b)
	return nil
}

func (f *CloudSpecForm) promptTopLevelFilter() error {
	if len(f.Fields.ResourceTypes) > 0 || len(f.Fields.NativeIDs) > 0 {
		return nil
	}

	const resourceTypeFilter = "Resource Type"
	const nativeIDFilter = "Native ID"
	prompt := selection.New("Filter resources by", []string{
		resourceTypeFilter,
		nativeIDFilter,
	})
	choice, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	switch choice {
	case nativeIDFilter:
		return f.promptNativeIDs()
	case resourceTypeFilter:
		return f.promptResourceTypes()
	}
	return nil
}

func (f *CloudSpecForm) promptNativeIDs() error {
	if len(f.Fields.ResourceTypes) > 0 {
		return nil
	}

	prompt := &multiplePrompt{
		prompt:  textinput.New("Native ID"),
		another: confirmation.New("Add another native ID?", confirmation.No),
	}
	nativeIDs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.NativeIDs = nativeIDs
	return nil
}

func (f *CloudSpecForm) promptResourceTypes() error {
	if len(f.Fields.ResourceTypes) > 0 {
		return nil
	}

	prompt := &multiplePrompt{
		prompt:  textinput.New("Resource type"),
		another: confirmation.New("Add another resource type?", confirmation.No),
	}
	resourceTypes, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.ResourceTypes = resourceTypes
	return f.promptAdditionalFilters()
}

func (f *CloudSpecForm) promptAdditionalFilters() error {
	if err := f.promptEnvironmentIDs(); err != nil {
		return err
	}
	return f.promptLocations()
}

func (f *CloudSpecForm) promptEnvironmentIDs() error {
	if len(f.Fields.EnvironmentIDs) > 0 {
		return nil
	}

	prompt := optionalPrompt[[]string]{
		enable: confirmation.New("Would you like to also filter by environment ID?", confirmation.No),
		prompt: &multiplePrompt{
			prompt:  textinput.New("Environment ID"),
			another: confirmation.New("Add another environment ID?", confirmation.No),
		},
	}
	environmentIDs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.EnvironmentIDs = environmentIDs
	return nil
}

func (f *CloudSpecForm) promptLocations() error {
	if len(f.Fields.EnvironmentIDs) > 0 {
		return nil
	}

	prompt := optionalPrompt[[]string]{
		enable: confirmation.New("Would you like to also filter by location?", confirmation.No),
		prompt: &multiplePrompt{
			prompt:  textinput.New("Location"),
			another: confirmation.New("Add another location?", confirmation.No),
		},
	}
	locations, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.Locations = locations
	return nil
}
