package linters

import (
	"strings"

	"regexp"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type repoLinter struct {
	Fs afero.Afero
}

func NewRepoLinter(fs afero.Afero) repoLinter {
	return repoLinter{fs}
}

func (r repoLinter) checkGlob(glob string) error {
	matches, err := afero.Glob(r.Fs, glob)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return errors.NewFileError(glob, "Could not find any files or directories matching glob")
	}
	return nil
}

func (r repoLinter) Lint(man manifest.Manifest) (result LintResult) {
	result.Linter = "Repo"

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

	for _, glob := range append(man.Repo.WatchedPaths, man.Repo.IgnoredPaths...) {
		if err := r.checkGlob(glob); err != nil {
			result.AddError(err)
		}
	}

	if man.Repo.GitCryptKey != "" && !regexp.MustCompile(`\(\([a-zA-Z-_]+\.[a-zA-Z-_]+\)\)`).MatchString(man.Repo.GitCryptKey) {
		result.AddError(errors.NewInvalidField("repo.git_crypt_key", "must be a vault secret"))
	}

	return
}
