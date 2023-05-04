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

package scaffold

import (
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-iac-rules/internal/project"
)

func checkProject(proj *project.Project, logger *zerolog.Logger) {
	// Test if we'll be able to query the project for Rule IDs and such
	_, err := proj.RuleMetadata()
	if err != nil {
		logger.Warn().Msgf("Found errors in this project. This tool is still usable, but we'll be unable to populate some menus: %s", err.Error())
	}
}
