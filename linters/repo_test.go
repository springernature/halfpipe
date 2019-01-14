package linters

import (
	"testing"

	"errors"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func testRepoLinter() repoLinter {
	return repoLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
		branchResolver: func() (branch string, err error) {
			return "master", nil
		},
		repoURIResolver: func() (string, error) {
			return "https://github.com/springernature/halfpipe.git", nil
		},
	}
}

func TestRepoIsEmpty(t *testing.T) {
	man := manifest.Manifest{}

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.uri", result.Errors[0])
}

func TestRepInvalidUri(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "goo"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.uri", result.Errors[0])
}

func TestRepoUriIsValidUri(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"

	result := testRepoLinter().Lint(man)
	assert.Empty(t, result.Errors)
	assert.Len(t, result.Warnings, 1)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "git@github.com:springernature/halfpipe.git"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "repo.private_key", result.Errors[0])

	man.Repo.PrivateKey = "somekey"
	result = testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestItChecksForWatchAndIgnores(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:          "https://github.com/springernature/halfpipe.git",
			BasePath:     "",
			WatchedPaths: []string{"watches/there", "watches/no-there/**"},
			IgnoredPaths: []string{"c/*", "d"},
		},
	}

	linter := testRepoLinter()
	linter.WorkingDir = "/repo"
	linter.Fs.Mkdir("/repo/watches/there", 0777)
	linter.Fs.Mkdir("/repo/c/d/e/f/g/h", 0777)

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 2)
}

func TestItChecksForWatchAndIgnoresRelativeToGitRoot(t *testing.T) {
	man := manifest.Manifest{
		Repo: manifest.Repo{
			URI:          "https://github.com/springernature/halfpipe.git",
			BasePath:     "project-name",
			WatchedPaths: []string{"watches/there", "watches/no-there/**"},
			IgnoredPaths: []string{"c/*", "d"},
		},
	}

	linter := testRepoLinter()
	linter.WorkingDir = "/home/projects/repo-project-name/project-name"
	linter.Fs.Mkdir("/home/projects/repo-project-name/watches/there", 0777)
	linter.Fs.Mkdir("/home/projects/repo-project-name/c/d/e/f/g/h", 0777)

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 2)
}

func TestRepoHasValidGitCryptKey(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.GitCryptKey = "((gitcrypt.key))"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRepoHasInvalidGitCryptKey(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.GitCryptKey = "CLEARTEXTKEY_BADASS"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.git_crypt_key", result.Errors[0])
}

func TestRepoWithPublicUrlAndPrivateKey(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.PrivateKey = "my_private_key"

	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.uri", result.Errors[0])
}

func TestRepoWhenBranchIsNotSetButOnMaster(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	result := testRepoLinter().Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRepoWhenBranchIsNotSetAndOnNonMasterBranch(t *testing.T) {
	currentBranch := "myBranch"
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return currentBranch, nil
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.branch", result.Errors[0])
}

func TestRepoWhenBranchIsSetAndOnNonMasterBranch(t *testing.T) {
	currentBranch := "myBranch"
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.Branch = currentBranch
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return currentBranch, nil
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestRepoWhenBranchIsSetToBranchXButYouAreOnY(t *testing.T) {
	currentBranch := "Y"
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.Branch = "X"
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return currentBranch, nil
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.branch", result.Errors[0])
}

func TestRepoWhenBranchIsSetToBranchXButYouAreOnMaster(t *testing.T) {
	currentBranch := "master"
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	man.Repo.Branch = "X"
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return currentBranch, nil
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidField(t, "repo.branch", result.Errors[0])
}

func TestRepoWhenBranchResolverReturnsError(t *testing.T) {
	expectedError := errors.New("Meeh")
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return "", expectedError
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, expectedError, result.Errors[0])
}

func TestRepoWhenRepoUriIsNotSameAsRepoResolver(t *testing.T) {
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	linter := testRepoLinter()
	linter.repoURIResolver = func() (string, error) {
		return "git@github.com:springernature/someRandomRepo.git", nil
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 2)
}

func TestPassesOnRepoUriResolverErrors(t *testing.T) {
	expectedError := errors.New("mehp")
	man := manifest.Manifest{}
	man.Repo.URI = "https://github.com/springernature/halfpipe.git"
	linter := testRepoLinter()
	linter.branchResolver = func() (branch string, err error) {
		return "master", nil
	}
	linter.repoURIResolver = func() (string, error) {
		return "", expectedError
	}

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, expectedError, result.Errors[0])
}
