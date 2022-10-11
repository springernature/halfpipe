package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestEmptyImage(t *testing.T) {
	trigger := manifest.DockerTrigger{}
	errors, warnings := LintDockerTrigger(trigger)

	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	assertContainsError(t, errors, NewErrMissingField("image"))
}

func TestOk(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "ubuntu",
	}
	errors, warnings := LintDockerTrigger(trigger)

	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}
