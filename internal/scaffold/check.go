package scaffold

import (
	"github.com/rs/zerolog"
	"github.com/snyk/cli-extension-cloud/internal/project"
)

func checkProject(proj *project.Project, logger *zerolog.Logger) {
	// Test if we'll be able to query the project for Rule IDs and such
	_, err := proj.RuleMetadata()
	if err != nil {
		logger.Warn().Msgf("Found errors in this project. This tool is still usable, but we'll be unable to populate some menus: %s", err.Error())
	}
}
