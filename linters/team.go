package linters

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type teamlinter struct{}

func NewTeamLinter() teamlinter {
	return teamlinter{}
}

func (teamlinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Team"
	result.DocsURL = "https://docs.halfpipe.io/docs/manifest/#team"

	if manifest.Team == "" {
		result.AddError(errors.NewMissingField("team"))
	}

	return
}
