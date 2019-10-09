package tasks

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDeployMLZipTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLZip{}

	errors, _ := LintDeployMLZipTask(task)

	if assert.Len(t, errors, 2) {
		linterrors.AssertMissingFieldInErrors(t, "target", errors)
		linterrors.AssertMissingFieldInErrors(t, "deploy_zip", errors)
	}
}

func TestDeployMLModulesTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLModules{}

	errors, _ := LintDeployMLModulesTask(task)

	if assert.Len(t, errors, 2) {
		linterrors.AssertMissingFieldInErrors(t, "target", errors)
		linterrors.AssertMissingFieldInErrors(t, "ml_modules_version", errors)
	}
}

func TestMLRetries(t *testing.T) {
	mlModule := manifest.DeployMLModules{}

	mlModule.Retries = -1
	errors, _ := LintDeployMLModulesTask(mlModule)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	mlModule.Retries = 6
	errors, _ = LintDeployMLModulesTask(mlModule)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	mlZip := manifest.DeployMLZip{}

	mlZip.Retries = -1
	errors, _ = LintDeployMLZipTask(mlZip)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	mlZip.Retries = 6
	errors, _ = LintDeployMLZipTask(mlZip)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)
}

func TestNotBothAppVersionAndUseBuildVersionAreSetMLModules(t *testing.T) {
	task := manifest.DeployMLModules{
		Targets:          []string{"localhost"},
		MLModulesVersion: "2.0",
		AppVersion:       "1.0",
		UseBuildVersion:  true,
	}

	errors, _ := LintDeployMLModulesTask(task)

	if assert.Len(t, errors, 1) {
		linterrors.AssertInvalidFieldInErrors(t, "use_build_version", errors)
	}
}

func TestNotBothAppVersionAndUseBuildVersionAreSetMLZip(t *testing.T) {
	task := manifest.DeployMLZip{
		Targets:         []string{"localhost"},
		AppVersion:      "1.0",
		DeployZip:       "foo.zip",
		UseBuildVersion: true,
	}

	errors, _ := LintDeployMLZipTask(task)

	if assert.Len(t, errors, 1) {
		linterrors.AssertInvalidFieldInErrors(t, "use_build_version", errors)
	}
}
