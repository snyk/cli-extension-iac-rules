package forms

import (
	"encoding/json"
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/snyk/cli-extension-cloud/internal/project"
)

type (
	RelationFields struct {
		Name                  string
		PrimaryResourceType   string
		PrimaryAttributes     []string
		SecondaryResourceType string
		SecondaryAttributes   []string
	}

	RelationForm struct {
		Project *project.Project
		Fields  RelationFields
	}
)

func (f *RelationForm) Run() error {
	if err := f.promptName(); err != nil {
		return err
	}
	if err := f.promptPrimaryResourceType(); err != nil {
		return err
	}
	if err := f.promptPrimaryAttributes(); err != nil {
		return err
	}
	if err := f.promptSecondaryResourceType(); err != nil {
		return err
	}
	if err := f.promptSecondaryAttributes(); err != nil {
		return err
	}

	leftAttributesJSON, err := json.Marshal(f.Fields.PrimaryAttributes)
	if err != nil {
		return err
	}
	rightAttributesJSON, err := json.Marshal(f.Fields.SecondaryAttributes)
	if err != nil {
		return err
	}
	relation, err := templateRelation(relationParams{
		Name:              f.Fields.Name,
		LeftResourceType:  f.Fields.PrimaryResourceType,
		LeftAttributes:    string(leftAttributesJSON),
		RightResourceType: f.Fields.SecondaryResourceType,
		RightAttributes:   string(rightAttributesJSON),
	})
	if err != nil {
		return err
	}
	return f.Project.AddRelation(relation)
}

func (f *RelationForm) promptName() error {
	if f.Fields.Name != "" {
		return nil
	}

	prompt := textinput.New("Relation name")
	name, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.Name = name
	return nil
}

func (f *RelationForm) promptPrimaryResourceType() error {
	if f.Fields.PrimaryResourceType != "" {
		return nil
	}

	prompt := textinput.New("Primary resource type")
	primary, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.PrimaryResourceType = primary
	return nil
}

func (f *RelationForm) promptPrimaryAttributes() error {
	if len(f.Fields.PrimaryAttributes) > 0 {
		return nil
	}

	prompt := &multiplePrompt{
		prompt:  textinput.New(fmt.Sprintf("Attribute from %s", f.Fields.PrimaryResourceType)),
		another: confirmation.New("Would you like to add another attribute?", confirmation.No),
	}
	attrs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.PrimaryAttributes = attrs
	return nil
}

func (f *RelationForm) promptSecondaryResourceType() error {
	if f.Fields.SecondaryResourceType != "" {
		return nil
	}

	prompt := textinput.New("Secondary resource type")
	secondary, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.SecondaryResourceType = secondary
	return nil
}

func (f *RelationForm) promptSecondaryAttributes() error {
	if len(f.Fields.SecondaryAttributes) > 0 {
		return nil
	}

	prompt := &multiplePrompt{
		prompt:  textinput.New(fmt.Sprintf("Attribute from %s", f.Fields.SecondaryResourceType)),
		another: confirmation.New("Would you like to add another attribute?", confirmation.No),
	}
	attrs, err := prompt.RunPrompt()
	if err != nil {
		return err
	}

	f.Fields.SecondaryAttributes = attrs
	return nil
}
