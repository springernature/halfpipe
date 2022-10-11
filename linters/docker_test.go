package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestEmptyImage(t *testing.T) {
	trigger := manifest.DockerTrigger{}
	errors := LintDockerTrigger(trigger)

	assertContainsError(t, errors, NewErrMissingField("image"))
}

func TestOk(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "ubuntu",
	}
	errors := LintDockerTrigger(trigger)

	assert.Len(t, errors, 0)
}
