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
