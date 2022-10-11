package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDeployMLZipTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLZip{}

	errors := LintDeployMLZipTask(task)

	if assert.Len(t, errors, 2) {
		assertContainsError(t, errors, NewErrMissingField("target"))
		assertContainsError(t, errors, NewErrMissingField("deploy_zip"))
	}
}

func TestDeployMLModulesTaskHasRequiredFields(t *testing.T) {
	task := manifest.DeployMLModules{}

	errors := LintDeployMLModulesTask(task)

	if assert.Len(t, errors, 2) {
		assertContainsError(t, errors, NewErrMissingField("target"))
		assertContainsError(t, errors, NewErrMissingField("ml_modules_version"))
	}
}

func TestMLRetries(t *testing.T) {
	mlModule := manifest.DeployMLModules{}

	mlModule.Retries = -1
	errors := LintDeployMLModulesTask(mlModule)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	mlModule.Retries = 6
	errors = LintDeployMLModulesTask(mlModule)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	mlZip := manifest.DeployMLZip{}

	mlZip.Retries = -1
	errors = LintDeployMLZipTask(mlZip)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	mlZip.Retries = 6
	errors = LintDeployMLZipTask(mlZip)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))
}

func TestNotBothAppVersionAndUseBuildVersionAreSetMLModules(t *testing.T) {
	task := manifest.DeployMLModules{
		Targets:          []string{"localhost"},
		MLModulesVersion: "2.0",
		AppVersion:       "1.0",
		UseBuildVersion:  true,
	}

	errors := LintDeployMLModulesTask(task)

	if assert.Len(t, errors, 1) {
		assertContainsError(t, errors, ErrInvalidField.WithValue("use_build_version"))
	}
}

func TestNotBothAppVersionAndUseBuildVersionAreSetMLZip(t *testing.T) {
	task := manifest.DeployMLZip{
		Targets:         []string{"localhost"},
		AppVersion:      "1.0",
		DeployZip:       "foo.zip",
		UseBuildVersion: true,
	}

	errors := LintDeployMLZipTask(task)

	if assert.Len(t, errors, 1) {
		assertContainsError(t, errors, ErrInvalidField.WithValue("use_build_version"))
	}
}
