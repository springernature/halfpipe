package linters

import (
	"testing"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var repoLinter = RepoLinter{}

func TestRepoIsEmpty(t *testing.T) {
	man := model.Manifest{}

	errs := repoLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.MissingField{}, errs[0])
}

func TestRepInvalidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "goo"

	errs := repoLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.InvalidField{}, errs[0])
}

func TestRepoUriIsValidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "https://github.com/springernature/halfpipe.git"

	errs := repoLinter.Lint(man)
	assert.Empty(t, errs)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:springernature/halfpipe.git"

	errs := repoLinter.Lint(manifest)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.MissingField{}, errs[0])

	manifest.Repo.PrivateKey = "somekey"
	errs = repoLinter.Lint(manifest)
	assert.Len(t, errs, 0)
}
