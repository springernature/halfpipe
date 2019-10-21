package tasks

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployCF{Manifest: "manifest.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, warnings := LintDeployCFTask(task, fs)
	assert.Len(t, errors, 4)
	assert.Len(t, warnings, 0)

	linterrors.AssertMissingFieldInErrors(t, "api", errors)
	linterrors.AssertMissingFieldInErrors(t, "space", errors)
	linterrors.AssertMissingFieldInErrors(t, "org", errors)
	linterrors.AssertFileErrorInErrors(t, "manifest.yml", errors)
}

func TestCFDeployTaskWithEmptyTestDomain(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:      "((cloudfoundry.api-dev))",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings := LintDeployCFTask(task, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)

	task = manifest.DeployCF{
		API:      "",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings = LintDeployCFTask(task, fs)
	assert.Len(t, errors, 1)
	linterrors.AssertMissingFieldInErrors(t, "api", errors)

	task = manifest.DeployCF{
		API:      "someRandomApi",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings = LintDeployCFTask(task, fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	linterrors.AssertMissingFieldInErrors(t, "testDomain", errors)

}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors, _ := LintDeployCFTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDeployCFTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}

func TestCFDeployTaskWithManifestFromArtifacts(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
	}

	errors, warnings := LintDeployCFTask(task, fs)

	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	linterrors.AssertFileErrorInErrors(t, "../artifacts/manifest.yml", warnings)
}

func TestCFDeployTaskWithManifestFromArtifactsAndPrePromoteShouldError(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
		PrePromote: []manifest.Task{
			manifest.Run{},
		},
	}

	errors, warnings := LintDeployCFTask(task, fs)

	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 1)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote", errors)
}

func TestCfPushPreStart(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		Manifest:   "../artifacts/manifest.yml",
		API:        "api",
		Space:      "space",
		Org:        "org",
		TestDomain: "test.domain",
	}

	task.PreStart = []string{"cf something good"}
	errors, _ := LintDeployCFTask(task, fs)
	assert.Empty(t, errors)

	task.PreStart = []string{"cf something good", "something bad", "cf something else good", "something else bad"}
	errors, _ = LintDeployCFTask(task, fs)
	assert.Len(t, errors, 2)
	linterrors.AssertInvalidFieldInErrors(t, "pre_start", errors)
}

func TestSubTasksDoesntDefineNotifications(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("manifest.yml", []byte("foo"), 0777)

	task := manifest.DeployCF{
		API:        "asdf",
		Space:      "asdf",
		Org:        "asdf",
		TestDomain: "asdf",
		PrePromote: manifest.TaskList{
			manifest.Run{Notifications: manifest.Notifications{OnSuccess: []string{"Meehp"}}},
			manifest.Run{},
			manifest.Run{Notifications: manifest.Notifications{OnFailure: []string{"Moohp"}}},
		},
	}

	errors, _ := LintDeployCFTask(task, fs)
	assert.Len(t, errors, 2)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote[0].notifications", errors)
	linterrors.AssertInvalidFieldInErrors(t, "pre_promote[2].notifications", errors)
}
