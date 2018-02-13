package linters

import (
	"strings"

	"github.com/springernature/halfpipe/model"
)

type RepoLinter struct{}

func (r RepoLinter) Lint(man model.Manifest) []error {
	var errs []error
	if man.Repo.Uri == "" {
		errs = append(errs, model.NewMissingField("repo.uri"))
		return errs
	}

	if !strings.HasSuffix(man.Repo.Uri, ".git") {
		errs = append(errs, model.NewInvalidField("repo.uri", "must end with .git"))
	}

	if strings.HasPrefix(man.Repo.Uri, "git@") && man.Repo.PrivateKey == "" {
		errs = append(errs, model.NewMissingField("repo.private_key"))
	}

	return errs
}
