package linters

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/parser"
)

type teamlinter struct{}

func NewTeamLinter() teamlinter {
	return teamlinter{}
}

func (teamlinter) Lint(manifest parser.Manifest) (result LintResult) {
	result.Linter = "Team"

	if manifest.Team == "" {
		result.AddError(errors.NewMissingField("team"))
	}

	return
}
