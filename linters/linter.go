package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

type Linter interface {
	Lint(manifest manifest.Manifest) LintResult
}
