package triggers

import (
	"errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

var defaultBranchResolver = func() (branch string, err error) {
	return "master", nil
}

var defaultRepoURIResolver = func(uri string) func() (string, error) {
	return func() (s string, e error) {
		return uri, nil
	}
}

func TestUriIsEmpty(t *testing.T) {
	trigger := manifest.GitTrigger{}

	errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

	assert.Len(t, errs, 1)
	helpers.AssertMissingFieldInErrors(t, "uri", errs)
}

func TestInvalidUri(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI: "goo",
	}

	errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

	assert.Len(t, errs, 1)
	helpers.AssertInvalidFieldInErrors(t, "uri", errs)
}

func TestUriIsValidHttpsUri(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI: "https://github.com/springernature/halfpipe.git",
	}

	errs, warns := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 1)
}

func TestPrivateRepoHasPrivateKeySet(t *testing.T) {
	t.Run("errors when private key is not set", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI: "git@github.com:springernature/halfpipe.git",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		helpers.AssertMissingFieldInErrors(t, "private_key", errs)
	})

	t.Run("no errors when private key is set", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/halfpipe.git",
			PrivateKey: "kehe",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 0)
	})
}

func TestItChecksForWatchAndIgnores(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:          "https://github.com/springernature/halfpipe.git",
		WatchedPaths: []string{"watches/there", "watches/no-there/**"},
		IgnoredPaths: []string{"c/*", "d"},
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	workingDir := "/repo"

	fs.Mkdir(path.Join(workingDir, "watches/there"), 0777)
	fs.Mkdir(path.Join(workingDir, "c/d/e/f/g/h"), 0777)

	errs, _ := LintGitTrigger(trigger, fs, workingDir, defaultBranchResolver, defaultRepoURIResolver(trigger.URI))
	assert.Len(t, errs, 2)
	helpers.AssertFileErrorInErrors(t, trigger.WatchedPaths[1], errs)
	helpers.AssertFileErrorInErrors(t, trigger.IgnoredPaths[1], errs)
}

func TestItChecksForWatchAndIgnoresRelativeToGitRoot(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:          "https://github.com/springernature/halfpipe.git",
		BasePath:     "project-name",
		WatchedPaths: []string{"watches/there", "watches/no-there/**"},
		IgnoredPaths: []string{"c/*", "d"},
	}

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	workingDir := "/home/projects/repo-project-name/project-name"
	fs.Mkdir("/home/projects/repo-project-name/watches/there", 0777)
	fs.Mkdir("/home/projects/repo-project-name/c/d/e/f/g/h", 0777)

	errs, _ := LintGitTrigger(trigger, fs, workingDir, defaultBranchResolver, defaultRepoURIResolver(trigger.URI))
	assert.Len(t, errs, 2)
	helpers.AssertFileErrorInErrors(t, trigger.WatchedPaths[1], errs)
	helpers.AssertFileErrorInErrors(t, trigger.IgnoredPaths[1], errs)
}

func TestHasValidGitCryptKey(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:         "git@github.com:springernature/halfpipe.git",
			PrivateKey:  "kehe",
			GitCryptKey: "((gitcrypt.key))",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 0)
	})

	t.Run("invalid", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:         "git@github.com:springernature/halfpipe.git",
			PrivateKey:  "kehe",
			GitCryptKey: "CLEARTEXTKEY_BADASS",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "git_crypt_key", errs)

	})
}

func TestPublicUrIAndPrivateKey(t *testing.T) {
	trigger := manifest.GitTrigger{
		URI:        "https://github.com/springernature/halfpipe.git",
		PrivateKey: "kehe",
	}

	errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, defaultRepoURIResolver(trigger.URI))

	assert.Len(t, errs, 1)
	helpers.AssertInvalidFieldInErrors(t, "uri", errs)
}

func TestBranch(t *testing.T) {

	t.Run("when branch is set to the same branch as we are on", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: currentBranch,
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 0)
	})

	t.Run("when branch is not set and on non-master branch", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI: "https://github.com/springernature/halfpipe.git",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "branch", errs)
	})

	t.Run("when branch is set to some other branch than we are on", func(t *testing.T) {
		currentBranch := "myBranch"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "branch", errs)
	})

	t.Run("when branch is set but we are on master", func(t *testing.T) {
		currentBranch := "master"
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return currentBranch, nil
		}, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "branch", errs)
	})

	t.Run("when branchResolver returns an error", func(t *testing.T) {
		expectedError := errors.New("kehe")
		trigger := manifest.GitTrigger{
			URI:    "https://github.com/springernature/halfpipe.git",
			Branch: "someRandomBranch",
		}

		errs, _ := LintGitTrigger(trigger, afero.Afero{}, "", func() (branch string, err error) {
			return "", expectedError
		}, defaultRepoURIResolver(trigger.URI))

		assert.Len(t, errs, 1)
		assert.Contains(t, errs, expectedError)
	})
}

func TestRepoResolver(t *testing.T) {
	t.Run("when uri is not same as uri resolver", func(t *testing.T) {
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/someRepo.git",
			PrivateKey: "asdf",
		}

		errs, warnings := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, func() (s string, e error) {
			return "git@github.com:springernature/someRandomRepo.git", nil
		})

		assert.Len(t, errs, 0)
		assert.Len(t, warnings, 1)
	})

	t.Run("passes on error from uri resolver", func(t *testing.T) {
		expectedError := errors.New("keHu")
		trigger := manifest.GitTrigger{
			URI:        "git@github.com:springernature/someRepo.git",
			PrivateKey: "asdf",
		}

		errs, warnings := LintGitTrigger(trigger, afero.Afero{}, "", defaultBranchResolver, func() (s string, e error) {
			return "", expectedError
		})

		assert.Len(t, warnings, 0)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs, expectedError)
	})
}
