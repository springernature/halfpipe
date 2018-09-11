package tasks

import (
	"testing"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/helpers"
	"github.com/stretchr/testify/assert"
)

func TestDeployMLZipTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLZip{}

	errors, _ := LintDeployMLZipTask(task, "mlModule")

	if assert.Len(t, errors, 2) {
		helpers.AssertMissingFieldInErrors(t, "deploy-ml.target", errors)
		helpers.AssertMissingFieldInErrors(t, "deploy-ml.deploy_zip", errors)
	}
}

func TestDeployMLModulesTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLModules{}

	errors, _ := LintDeployMLModulesTask(task, "mlModule")

	if assert.Len(t, errors, 2) {
		helpers.AssertMissingFieldInErrors(t, "deploy-ml.target", errors)
		helpers.AssertMissingFieldInErrors(t, "deploy-ml.ml_modules_version", errors)
	}
}

func TestMLRetries(t *testing.T) {
	mlModule := manifest.DeployMLModules{}

	mlModule.Retries = -1
	errors, _ := LintDeployMLModulesTask(mlModule, "mlModule")
	helpers.AssertInvalidFieldInErrors(t, "deploy-ml-modules.retries", errors)

	mlModule.Retries = 6
	errors, _ = LintDeployMLModulesTask(mlModule, "mlModule")
	helpers.AssertInvalidFieldInErrors(t, "deploy-ml-modules.retries", errors)

	mlZip := manifest.DeployMLZip{}

	mlZip.Retries = -1
	errors, _ = LintDeployMLZipTask(mlZip, "mlZip")
	helpers.AssertInvalidFieldInErrors(t, "deploy-ml-zip.retries", errors)

	mlZip.Retries = 6
	errors, _ = LintDeployMLZipTask(mlZip, "mlZip")
	helpers.AssertInvalidFieldInErrors(t, "deploy-ml-zip.retries", errors)
}
