package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestObservabilityProvisioningTaskWhenSpecifiedFolderDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintObservabilityProvisioningTask(manifest.ObservabilityProvisioning{SLOs: "slos"}, fs)

	assertContainsError(t, errors, NewErrFolderNotFound("slos"))
}

func TestObservabilityProvisioningTaskWhenSpecifiedFolderExists(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintObservabilityProvisioningTask(manifest.ObservabilityProvisioning{}, fs)

	assert.Empty(t, errors, "there should be no errors")
}

// TODO: Add test to check if there not parameters for the task, to avoid having a task doing nothing.
