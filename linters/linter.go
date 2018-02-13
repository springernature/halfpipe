package linters

import "github.com/springernature/halfpipe/model"

type Linter interface {
	Lint(manifest model.Manifest) []error
}
