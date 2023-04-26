package forms

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/open-policy-agent/opa/format"
)

//go:embed ruletemplates/multi.rego.tmpl
var multiRegoTmpl string

//go:embed ruletemplates/single.rego.tmpl
var singleRegoTmpl string

//go:embed ruletemplates/relation.rego.tmpl
var relationRegoTmpl string

var multiResourceRuleTemplate = template.Must(
	template.New("MultiResourceRule").Parse(multiRegoTmpl))

var singleResourceRuleTemplate = template.Must(
	template.New("SingleResourceRule").Parse(singleRegoTmpl))

var relationTemplate = template.Must(
	template.New("Relation").Parse(relationRegoTmpl))

type multiResourceRuleParams struct {
	RulePackage               string
	InputType                 string
	RuleMetadata              string
	PrimaryResourceType       string
	PrimaryResourcePlural     string
	PrimaryResourceSingular   string
	SecondaryResourcePlural   string
	SecondaryResourceSingular string
	Relation                  string
}

func templateMultiResourceRule(params multiResourceRuleParams) ([]byte, error) {
	var buf bytes.Buffer
	err := multiResourceRuleTemplate.Execute(&buf, params)
	if err != nil {
		return nil, err
	}
	return format.Source("", buf.Bytes())
}

type singleResourceRuleParams struct {
	RulePackage  string
	InputType    string
	RuleMetadata string
	ResourceType string
}

func templateSingleResourceRule(params singleResourceRuleParams) ([]byte, error) {
	var buf bytes.Buffer
	err := singleResourceRuleTemplate.Execute(&buf, params)
	if err != nil {
		return nil, err
	}
	return format.Source("", buf.Bytes())
}

type relationParams struct {
	Name              string
	LeftResourceType  string
	LeftAttributes    []string
	RightResourceType string
	RightAttributes   []string
}

func templateRelation(params relationParams) (string, error) {
	var buf bytes.Buffer
	err := relationTemplate.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
