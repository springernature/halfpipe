package linters

import (
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type TeamLinter struct{}

func (TeamLinter) Lint(manifest model.Manifest) (result errors.LintResult) {
	result.Linter = "Team Linter"

	if manifest.Team == "" {
		result.Errors = append(result.Errors, errors.NewMissingField("team"))
	}

	return
}
