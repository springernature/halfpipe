package triggers

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptyImage(t *testing.T) {
	trigger := manifest.DockerTrigger{}
	errs, warns := LintDockerTrigger(trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	helpers.AssertMissingFieldInErrors(t, "image", errs)
}

func TestOk(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "ubuntu",
	}
	errs, warns := LintDockerTrigger(trigger)

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 0)
}
