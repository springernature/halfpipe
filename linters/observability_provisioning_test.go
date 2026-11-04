package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestObservabilityProvisioningTask_WhenSLOFolderDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintObservabilityProvisioningTask(manifest.ObservabilityProvisioning{SLOs: "slos"}, fs)

	assertContainsError(t, errors, NewErrFolderNotFound("slos"))
	assert.Len(t, errors, 1, "there should be one error")
}

func TestObservabilityProvisioningTask_WhenMetricRulesFolderDoesNotExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors := LintObservabilityProvisioningTask(manifest.ObservabilityProvisioning{MetricRules: "metric_rules"}, fs)

	assertContainsError(t, errors, NewErrFolderNotFound("metric_rules"))
	assert.Len(t, errors, 1, "there should be one error")
}

// TODO: Add test to check if there not parameters for the task, to avoid having a task doing nothing.
