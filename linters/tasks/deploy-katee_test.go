package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKateeDeployTaskWithEmptyTask(t *testing.T) {
	task := manifest.DeployKatee{VelaAppFile: "vela.yml"}
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, warnings := LintDeployKateeTask(task, fs)
	assert.Len(t, errors, 2)
	assert.Len(t, warnings, 0)

	linterrors.AssertMissingFieldInErrors(t, "applicationName", errors)
	linterrors.AssertFileErrorInErrors(t, "vela.yml", errors)
}

func TestKateeDeployRetries(t *testing.T) {
	task := manifest.DeployKatee{}

	task.Retries = -1
	errors, _ := LintDeployKateeTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDeployKateeTask(task, afero.Afero{Fs: afero.NewMemMapFs()})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}
