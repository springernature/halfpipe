package linters

import (
	"strings"

	"github.com/springernature/halfpipe/model"
)

type Repo struct{}

func (r Repo) Lint(man model.Manifest) []error {
	var errs []error
	if man.Repo.Uri == "" {
		errs = append(errs, model.NewMissingField("repo.uri"))
		return errs
	}

	if !strings.HasSuffix(man.Repo.Uri, ".git") {
		errs = append(errs, model.NewInvalidField("repo.uri", "must end with .git"))
	}
	return errs
}
