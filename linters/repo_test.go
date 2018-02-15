package linters

import (
	"testing"

	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var repoLinter = RepoLinter{}

func TestRepoIsEmpty(t *testing.T) {
	man := model.Manifest{}

	result := repoLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.uri", result.Errors[0])
}

func TestRepInvalidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "goo"

	result := repoLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.uri", result.Errors[0])
}

func TestRepoUriIsValidUri(t *testing.T) {
	man := model.Manifest{}
	man.Repo.Uri = "https://github.com/springernature/halfpipe.git"

	result := repoLinter.Lint(man)
	assert.Empty(t, result.Errors)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	manifest := model.Manifest{}
	manifest.Repo.Uri = "git@github.com:springernature/halfpipe.git"

	result := repoLinter.Lint(manifest)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.private_key", result.Errors[0])

	manifest.Repo.PrivateKey = "somekey"
	result = repoLinter.Lint(manifest)
	assert.Len(t, result.Errors, 0)
}
