package linters

import (
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/errors"
)

type Linter interface {
	Lint(manifest model.Manifest) errors.LintResult
}
