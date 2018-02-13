package linters

import (
	"strings"

	"regexp"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type RepoLinter struct{}

func (r RepoLinter) Lint(man model.Manifest) []error {
	var errs []error
	if man.Repo.Uri == "" {
		errs = append(errs, errors.NewMissingField("repo.uri"))
		return errs
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, man.Repo.Uri)
	if !match {
		errs = append(errs, errors.NewInvalidField("repo.uri", "must be a valid git uri"))
	}

	if strings.HasPrefix(man.Repo.Uri, "git@") && man.Repo.PrivateKey == "" {
		errs = append(errs, errors.NewMissingField("repo.private_key"))
	}

	return errs
}
