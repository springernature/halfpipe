package triggers

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestEmptyImage(t *testing.T) {
	trigger := manifest.DockerTrigger{}
	errs, warns := LintDockerTrigger(trigger)

	assert.Len(t, errs, 1)
	assert.Len(t, warns, 0)
	linterrors.AssertMissingFieldInErrors(t, "image", errs)
}

func TestOk(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "ubuntu",
	}
	errs, warns := LintDockerTrigger(trigger)

	assert.Len(t, errs, 0)
	assert.Len(t, warns, 0)
}
