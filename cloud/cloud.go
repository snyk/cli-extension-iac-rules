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

package cloud

import (
	"github.com/snyk/cli-extension-cloud/internal/scaffold"
	"github.com/snyk/go-application-framework/pkg/workflow"

	"github.com/snyk/cli-extension-cloud/internal/push"
	"github.com/snyk/cli-extension-cloud/internal/spec"
)

func Init(e workflow.Engine) error {
	if err := scaffold.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := spec.RegisterWorkflows(e); err != nil {
		return err
	}
	if err := push.RegisterWorkflows(e); err != nil {
		return err
	}
	return nil
}
