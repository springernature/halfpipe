package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/parser"
	"github.com/stretchr/testify/assert"
)

func testRepoLinter() RepoLinter {
	return RepoLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestRepoIsEmpty(t *testing.T) {
	man := parser.Manifest{}

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.uri", result.Errors[0])
}

func TestRepInvalidUri(t *testing.T) {
	man := parser.Manifest{}
	man.Repo.Uri = "goo"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.uri", result.Errors[0])
}

func TestRepoUriIsValidUri(t *testing.T) {
	man := parser.Manifest{}
	man.Repo.Uri = "https://github.com/springernature/halfpipe.git"

	result := testRepoLinter().Lint(man)
	assert.Empty(t, result.Errors)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	manifest := parser.Manifest{}
	manifest.Repo.Uri = "git@github.com:springernature/halfpipe.git"

	result := testRepoLinter().Lint(manifest)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.private_key", result.Errors[0])

	manifest.Repo.PrivateKey = "somekey"
	result = testRepoLinter().Lint(manifest)
	assert.Len(t, result.Errors, 0)
}

func TestItChecksForWatchAndIgnores(t *testing.T) {
	watches := []string{"watches/there", "watches/no-there/**"}
	ignores := []string{"c/*", "d"}
	manifest := parser.Manifest{}
	manifest.Repo.Uri = "https://github.com/springernature/halfpipe.git"
	manifest.Repo.WatchedPaths = watches
	manifest.Repo.IgnoredPaths = ignores

	linter := testRepoLinter()
	linter.Fs.Mkdir("watches/there", 0777)
	linter.Fs.Mkdir("c/d/e/f/g/h", 0777)

	result := linter.Lint(manifest)
	assert.Len(t, result.Errors, 2)
}

func TestRepoHasValidGitCryptKey(t *testing.T) {
	manifest := parser.Manifest{}
	manifest.Repo.Uri = "https://github.com/springernature/halfpipe.git"
	manifest.Repo.GitCryptKey = "((gitcrypt.key))"

	result := testRepoLinter().Lint(manifest)
	assert.Len(t, result.Errors, 0)
}

func TestRepoHasInvalidGitCryptKey(t *testing.T) {
	manifest := parser.Manifest{}
	manifest.Repo.Uri = "https://github.com/springernature/halfpipe.git"
	manifest.Repo.GitCryptKey = "CLEARTEXTKEY_BADASS"

	result := testRepoLinter().Lint(manifest)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.git_crypt_key", result.Errors[0])
}
