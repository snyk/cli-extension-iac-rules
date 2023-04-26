package forms

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed ruletemplates/relation.rego.tmpl
var relationRegoTmpl string

var relationTemplate = template.Must(
	template.New("Relation").Parse(relationRegoTmpl))

type relationParams struct {
	Name              string
	LeftResourceType  string
	LeftAttributes    string
	RightResourceType string
	RightAttributes   string
}

func templateRelation(params relationParams) (string, error) {
	var buf bytes.Buffer
	err := relationTemplate.Execute(&buf, params)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
