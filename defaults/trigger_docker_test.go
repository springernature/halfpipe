package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestDockerTrigger(t *testing.T) {
	t.Run("does not do anything when the image is not from our registry", func(t *testing.T) {
		trigger := manifest.DockerTrigger{
			Image: "ubuntu",
		}

		assert.Equal(t, trigger, defaultDockerTrigger(trigger, DefaultValues))
	})

	t.Run("sets the username and password if not set when using private registry", func(t *testing.T) {
		trigger := manifest.DockerTrigger{
			Image: path.Join(config.DockerRegistry, "ubuntu"),
		}

		expectedTrigger := manifest.DockerTrigger{
			Image:    path.Join(config.DockerRegistry, "ubuntu"),
			Username: DefaultValues.DockerUsername,
			Password: DefaultValues.DockerPassword,
		}

		assert.Equal(t, expectedTrigger, defaultDockerTrigger(trigger, DefaultValues))
	})
}
