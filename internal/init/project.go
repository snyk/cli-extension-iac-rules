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

package init

import (
	"os"
	"path/filepath"

	"github.com/snyk/cli-extension-iac-rules/internal/init/forms"
	"github.com/snyk/cli-extension-iac-rules/internal/project"
	"github.com/snyk/go-application-framework/pkg/workflow"
	"github.com/spf13/afero"
)

func ProjectWorkflow(
	ictx workflow.InvocationContext,
	_ []workflow.Data,
) ([]workflow.Data, error) {
	logger := ictx.GetEnhancedLogger()
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	proj, err := project.FromDir(afero.NewOsFs(), ".")
	if err != nil {
		return nil, err
	}
	defaultName := filepath.Base(wd)
	if name := proj.Manifest().Name; name != "" {
		defaultName = name
	}
	form := &forms.ProjectForm{
		Project:     proj,
		DefaultName: defaultName,
		Logger:      logger,
	}
	if err := form.Run(); err != nil {
		return nil, err
	}
	if err := proj.WriteChanges(); err != nil {
		return nil, err
	}
	return []workflow.Data{}, nil
}
