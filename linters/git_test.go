package linters

import (
	"errors"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var defaultBranchResolver = func() (branch string, err error) {
	return "main", nil
}

var defaultRepoURIResolver = func(uri string) func() (string, error) {
	return func() (s string, e error) {
		return uri, nil
	}
}

func TestUriIsEmpty(t *testing.T) {
	trigger := manifest.GitTrigger{}

	errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

	assert.Len(t, errs, 1)
	assertContainsError(t, errs, NewErrMissingField("uri"))
}

func TestInvalidUri(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI: "goo",
	}

	errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

	assert.Len(t, errs, 1)
	assertContainsError(t, errs, ErrInvalidField.WithValue("uri"))
}

func TestUriIsValidHttpsUri(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:    "https://github.com/springernature/halfpipe.git",
		Branch: "main",
	}

	errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

	assertContainsError(t, errs, ErrInvalidField.WithValue("uri"))
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	t.Run("errors when private key is not set", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:    "git@github.com:springernature/halfpipe.git",
			Branch: "main",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assert.Len(t, errs, 1)
		assertContainsError(t, errs, NewErrMissingField("private_key"))
	})

	t.Run("no errors when private key is set", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/halfpipe.git",
			PrivateKey: "kehe",
			Branch:     "main",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assert.Len(t, errs, 0)
	})
}

func TestItChecksForWatchedPaths(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:          "https://github.com/springernature/halfpipe.git",
		WatchedPaths: []string{"watches/there", "watches/no-there/**"},
		IgnoredPaths: []string{"c/*", "d"},
		Branch:       "main",
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	workingDir := "/repo"

	fs.Mkdir(path.Join(workingDir, "watches/there"), 0777)

	errs := LintGitTrigger(trigger, fs, workingDir, defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))
	assertContainsError(t, errs, ErrFileNotFound)
}

func TestItChecksForWatchedPathsRelativeToGitRoot(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:          "https://github.com/springernature/halfpipe.git",
		Branch:       "main",
		BasePath:     "project-name",
		WatchedPaths: []string{"watches/there", "watches/no-there/**"},
		IgnoredPaths: []string{"c/*", "d"},
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	workingDir := "/home/projects/repo-project-name/project-name"
	fs.Mkdir("/home/projects/repo-project-name/watches/there", 0777)

	errs := LintGitTrigger(trigger, fs, workingDir, defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))
	assertContainsError(t, errs, ErrFileNotFound)
}

func TestHasValidGitCryptKey(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:         "git@github.com:springernature/halfpipe.git",
			PrivateKey:  "kehe",
			Branch:      "main",
			GitCryptKey: "((gitcrypt.key))",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assert.Empty(t, errs)
	})

	t.Run("invalid", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:         "git@github.com:springernature/halfpipe.git",
			Branch:      "main",
			PrivateKey:  "kehe",
			GitCryptKey: "CLEARTEXTKEY_BADASS",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assert.Len(t, errs, 1)
		assertContainsError(t, errs, ErrInvalidField.WithValue("git_crypt_key"))

	})
}

func TestPublicUrIAndPrivateKey(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:        "https://github.com/springernature/halfpipe.git",
		Branch:     "main",
		PrivateKey: "kehe",
	}

	errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))
	assertContainsError(t, errs, ErrInvalidField.WithValue("uri"))
}

func TestBranch(t *testing.T) {

	t.Run("when branch is set to the same branch as we are on", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: currentBranch,
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assertNotContainsError(t, errs, ErrInvalidField.WithValue("branch"))
	})

	t.Run("when branch is not set and on non-main branch", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI: "https://github.com/springernature/halfpipe.git",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assertContainsError(t, errs, ErrInvalidField.WithValue("branch"))
	})

	t.Run("when branch is set to some other branch than we are on", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assertContainsError(t, errs, ErrInvalidField.WithValue("branch"))
	})

	t.Run("when branch is set but we are on main", func(t *testing.T) {
		currentBranch := "main"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assertContainsError(t, errs, ErrInvalidField.WithValue("branch"))
	})

	t.Run("when branchResolver returns an error", func(t *testing.T) {
		expectedError := errors.New("kehe")
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return "", expectedError
		}, defaultRepoURIResolver(trigger.URI), manifest.Platform("concourse"))

		assert.Contains(t, errs, NewErrExternal(expectedError).AsWarning())
	})
}

func TestRepoResolver(t *testing.T) {
	t.Run("when uri is not same as uri resolver", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/someRepo.git",
			Branch:     "main",
			PrivateKey: "asdf",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, func() (s string, e error) {
			return "git@github.com:springernature/someRandomRepo.git", nil
		}, manifest.Platform("concourse"))

		assertContainsError(t, errs, ErrInvalidField.WithValue("uri"))
	})

	t.Run("passes on error from uri resolver", func(t *testing.T) {
		expectedError := errors.New("keHu")
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/someRepo.git",
			Branch:     "main",
			PrivateKey: "asdf",
		}

		errs := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, func() (s string, e error) {
			return "", expectedError
		}, manifest.Platform("concourse"))

		assert.Equal(t, errs[0], NewErrExternal(expectedError).AsWarning())
	})
}
