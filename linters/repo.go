package linters

import (
	"strings"

	"github.com/springernature/halfpipe/model"
	"regexp"
)

type RepoLinter struct{}

func (r RepoLinter) Lint(man model.Manifest) []error {
	var errs []error
	if man.Repo.Uri == "" {
		errs = append(errs, model.NewMissingField("repo.uri"))
		return errs
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, man.Repo.Uri)
	if !match {
		errs = append(errs, model.NewInvalidField("repo.uri", "must be a valid git uri"))
	}

	if strings.HasPrefix(man.Repo.Uri, "git@") && man.Repo.PrivateKey == "" {
		errs = append(errs, model.NewMissingField("repo.private_key"))
	}

	return errs
}
