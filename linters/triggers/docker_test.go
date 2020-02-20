package triggers

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestEmptyImage(t *testing.T) {
	trigger := manifest.DockerTrigger{}
	errors, warnings := LintDockerTrigger(trigger, []string{})

	assert.Len(t, errors, 1)
	assert.Len(t, warnings, 0)
	linterrors.AssertMissingFieldInErrors(t, "image", errors)
}

func TestOk(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "ubuntu",
	}
	errors, warnings := LintDockerTrigger(trigger, []string{})

	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestDeprecatedRegistry(t *testing.T) {
	trigger := manifest.DockerTrigger{
		Image: "deprecated.registry/image",
	}

	errors, warnings := LintDockerTrigger(trigger, []string{"foo.bar", "deprecated.registry"})

	assert.Len(t, errors, 0)
	if assert.Len(t, warnings, 1) {
		assert.Equal(t, linterrors.NewDeprecatedDockerRegistryError("deprecated.registry"), warnings[0])
	}

}
