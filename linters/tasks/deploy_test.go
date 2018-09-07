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

	errors, warnings := LintDeployCFTask(task, "task", fs)
	assert.Len(t, errors, 4)
	assert.Len(t, warnings, 0)

	helpers.AssertMissingFieldInErrors(t, "task deploy-cf.api", errors)
	helpers.AssertMissingFieldInErrors(t, "task deploy-cf.space", errors)
	helpers.AssertMissingFieldInErrors(t, "task deploy-cf.org", errors)
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

	errors, warnings := LintDeployCFTask(task, "task", fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)

	task = manifest.DeployCF{
		API:      "",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings = LintDeployCFTask(task, "task", fs)
	assert.Len(t, errors, 1)
	helpers.AssertMissingFieldInErrors(t, "task deploy-cf.api", errors)

	task = manifest.DeployCF{
		API:      "someRandomApi",
		Org:      "Something",
		Space:    "Something",
		Manifest: "manifest.yml",
	}

	errors, warnings = LintDeployCFTask(task, "task", fs)
	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	helpers.AssertMissingFieldInErrors(t, "task deploy-cf.testDomain", errors)

}
