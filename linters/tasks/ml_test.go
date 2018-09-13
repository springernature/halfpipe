package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeployMLZipTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLZip{}

	errors, _ := LintDeployMLZipTask(task)

	if assert.Len(t, errors, 2) {
		helpers.AssertMissingFieldInErrors(t, "target", errors)
		helpers.AssertMissingFieldInErrors(t, "deploy_zip", errors)
	}
}

func TestDeployMLModulesTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLModules{}

	errors, _ := LintDeployMLModulesTask(task)

	if assert.Len(t, errors, 2) {
		helpers.AssertMissingFieldInErrors(t, "target", errors)
		helpers.AssertMissingFieldInErrors(t, "ml_modules_version", errors)
	}
}

func TestMLRetries(t *testing.T) {
	mlModule := manifest.DeployMLModules{}

	mlModule.Retries = -1
	errors, _ := LintDeployMLModulesTask(mlModule)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	mlModule.Retries = 6
	errors, _ = LintDeployMLModulesTask(mlModule)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	mlZip := manifest.DeployMLZip{}

	mlZip.Retries = -1
	errors, _ = LintDeployMLZipTask(mlZip)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	mlZip.Retries = 6
	errors, _ = LintDeployMLZipTask(mlZip)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)
}
