package linters

import (
	"strings"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type teamlinter struct{}

func NewTeamLinter() teamlinter {
	return teamlinter{}
}

func (teamlinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Manifest"
	result.DocsURL = "https://docs.halfpipe.io/docs/manifest/"

	if manifest.Team == "" {
		result.AddError(errors.NewMissingField("team"))
	}

	if manifest.Pipeline == "" {
		result.AddError(errors.NewMissingField("pipeline"))
	}

	if strings.Contains(manifest.Pipeline, " ") {
		result.AddError(errors.NewInvalidField("pipeline", "pipeline name must not contains spaces!"))
	}

	return
}
