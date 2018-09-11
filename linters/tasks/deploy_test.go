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

func TestLintsSubTasksInDeployCF(t *testing.T) {
	task := manifest.DeployCF{
		API:   "((cloudfoundry.api-dev))",
		Org:   "org",
		Space: "space",
		PrePromote: []manifest.Task{
			manifest.Run{
				ManualTrigger: true,
				Parallel:      true,
			},
			manifest.DockerCompose{
				ManualTrigger: true,
				Parallel:      true,
			},
			manifest.DeployCF{
				ManualTrigger: true,
				Parallel:      true,
			},
			manifest.DockerPush{
				ManualTrigger: true,
				Parallel:      true,
			},
		},
	}

	errors, _ := LintDeployCFTask(task, "task", afero.Afero{Fs: afero.NewMemMapFs()})

	helpers.AssertInvalidFieldInErrors(t, "pre_promote[0] run.manual_trigger", errors)
	helpers.AssertInvalidFieldInErrors(t, "pre_promote[0] run.parallel", errors)

	helpers.AssertInvalidFieldInErrors(t, "pre_promote[1] docker-compose.manual_trigger", errors)
	helpers.AssertInvalidFieldInErrors(t, "pre_promote[1] docker-compose.parallel", errors)

	helpers.AssertInvalidFieldInErrors(t, "pre_promote[2] run.type", errors)
	helpers.AssertInvalidFieldInErrors(t, "pre_promote[3] run.type", errors)
}

func TestCfPushRetries(t *testing.T) {
	task := manifest.DeployCF{}

	task.Retries = -1
	errors, _ := LintDeployCFTask(task, "task", afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "deploy-cf.retries", errors)

	task.Retries = 6
	errors, _ = LintDeployCFTask(task, "task", afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "deploy-cf.retries", errors)
}

func TestLintsTheTimeoutInDeployTask(t *testing.T) {
	errors, _ := LintDeployCFTask(manifest.DeployCF{Timeout: "notAValidDuration"}, "task", afero.Afero{Fs: afero.NewMemMapFs()})
	helpers.AssertInvalidFieldInErrors(t, "deploy-cf.timeout", errors)
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

	errors, warnings := LintDeployCFTask(task, "task", fs)

	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 1)
	helpers.AssertFileErrorInErrors(t,"../artifacts/manifest.yml", warnings)
}
