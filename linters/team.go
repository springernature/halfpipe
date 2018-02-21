package linters

import (
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type TeamLinter struct{}

func (TeamLinter) Lint(manifest model.Manifest) (result model.LintResult) {
	result.Linter = "Team Linter"

	if manifest.Team == "" {
		result.AddError(errors.NewMissingField("team"))
	}

	return
}
