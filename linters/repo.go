package linters

import (
	"strings"

	"regexp"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type RepoLinter struct{}

func (r RepoLinter) Lint(man model.Manifest) (result errors.LintResult) {
	result.Linter = "Repo Linter"

	if man.Repo.Uri == "" {
		result.AddError(errors.NewMissingField("repo.uri"))
		return
	}

	match, _ := regexp.MatchString(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`, man.Repo.Uri)
	if !match {
		result.AddError(errors.NewInvalidField("repo.uri", "must be a valid git uri"))
		return
	}

	if strings.HasPrefix(man.Repo.Uri, "git@") && man.Repo.PrivateKey == "" {
		result.AddError(errors.NewMissingField("repo.private_key"))
	}

	return
}
