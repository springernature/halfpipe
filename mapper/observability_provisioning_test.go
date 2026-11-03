package mapper

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestObservabilityProvisioningMapper_ReturnsRunTask(t *testing.T) {
	mapper := NewObservabilityProvisioningMapper()

	updated, err := mapper.Apply(manifest.Manifest{Tasks: manifest.TaskList{manifest.ObservabilityProvisioning{
		SLOs:        "my-slos",
		MetricRules: "my-metric-rules",
		Name:        "my-name",
	}.SetVars(manifest.Vars{"BOB": "BEN"})}})

	assert.NoError(t, err)
	run := updated.Tasks[0].(manifest.Run)
	assert.Equal(t, "eu.gcr.io/halfpipe-io/engineering-enablement/o11ytool:latest", run.Docker.Image, "wrong image")
	assert.NotEmpty(t, run.Docker.Username)
	assert.NotEmpty(t, run.Docker.Password)
	assert.Equal(t, `\upload-slos -i my-slos && \sync-metric-rules -i my-metric-rules`, run.Script, "wrong script")
	assert.Equal(t, "my-name", run.Name, "wrong name")
	assert.Equal(t, manifest.Vars{"BOB": "BEN"}, run.Vars, "wrong vars")
}
