package forms

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/snyk/policy-engine/pkg/input"
)

//go:embed spectemplates/arm.json
var armTmpl []byte

//go:embed spectemplates/cfn.yaml
var cfnTmpl []byte

//go:embed spectemplates/k8s.yaml
var k8sTmpl []byte

//go:embed spectemplates/infra.tf
var tfTmpl []byte

func specForInputType(inputType string, name string) (filename string, contents []byte) {
	switch inputType {
	case input.Terraform.Name:
		filename = addExtIfNeeded(name, ".tf")
		contents = tfTmpl
	case input.Kubernetes.Name:
		filename = addExtIfNeeded(name, ".yaml")
		contents = k8sTmpl
	case input.CloudFormation.Name:
		filename = addExtIfNeeded(name, ".yaml")
		contents = cfnTmpl
	case input.Arm.Name:
		filename = addExtIfNeeded(name, ".json")
		contents = armTmpl
	}
	return
}

func addExtIfNeeded(name, ext string) string {
	if filepath.Ext(name) != ext {
		return fmt.Sprintf("%s%s", name, ext)
	}
	return name
}
