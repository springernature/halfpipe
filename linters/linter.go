package linters

import (
	"github.com/springernature/halfpipe/parser"
)

type Linter interface {
	Lint(manifest parser.Manifest) LintResult
}
