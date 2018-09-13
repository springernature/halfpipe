package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployCF{Manifest: "manifest.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, warnings := LintDeployCFTask(task, fs)
	assert.Len(t, errors, 4)
	assert.Len(t, warnings, 0)

	helpers.AssertMissingFieldInErrors(t, "api", errors)
	helpers.AssertMissingFieldInErrors(t, "space", errors)
	helpers.AssertMissingFieldInErrors(t, "org", errors)
	helpers.AssertFileErrorInErrors(t, "manifest.yml", errors)
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
	helpers.AssertMissingFieldInErrors(t, "api", errors)

	task = manifest.DeployCF{
		API:      "someRandomApi",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings = LintDeployCFTask(task, fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingFieldInErrors(t, "testDomain", errors)

}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors, _ := LintDeployCFTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDeployCFTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)
}

func TestLintsTheTimeoutInDeployTask(t *testing.T) {
	errors, _ := LintDeployCFTask(manifest.DeployCF{Timeout: "notAValidDuration"}, afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "timeout", errors)
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
	helpers.AssertFileErrorInErrors(t, "../artifacts/manifest.yml", warnings)
}
