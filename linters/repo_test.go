package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var repoLinter = Repo{}

func TestRepoIsEmpty(t *testing.T) {
	man := model.Manifest{}

	errs := repoLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, model.MissingField{}, errs[0])
}

func TestRepInvalidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "goo"

	errs := repoLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, model.InvalidField{}, errs[0])
}

func TestRepoUriIsValidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "https://avlalbalba/halfpipe.git"

	errs := repoLinter.Lint(man)
	assert.Empty(t, errs)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@avlalbalba:halfpipe.git"

	errs := repoLinter.Lint(manifest)
	assert.Len(t, errs, 1)
	assert.IsType(t, model.MissingField{}, errs[0])

	manifest.Repo.PrivateKey = "somekey"
	errs = repoLinter.Lint(manifest)
	assert.Len(t, errs, 0)
}
