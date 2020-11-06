package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunTaskDockerDefault(t *testing.T) {

	t.Run("public docker image", func(t *testing.T) {
		task := manifest.Run{
			Script: "./blah",
			Docker: manifest.Docker{
				Image: "Blah",
			},
		}

		updated := runDefaulter(task, DefaultValues)
		assert.Equal(t, task, updated)
	})

	t.Run("with private docker image", func(t *testing.T) {
		task := manifest.Run{
			Script: "./blah",
			Docker: manifest.Docker{
				Image: config.DockerRegistry + "runImage",
			},
		}

		expectedTask := manifest.Run{
			Script: "./blah",
			Docker: manifest.Docker{
				Image:    config.DockerRegistry + "runImage",
				Username: DefaultValues.Docker.Username,
				Password: DefaultValues.Docker.Password,
			},
		}

		assert.Equal(t, expectedTask, runDefaulter(task, DefaultValues))
	})

}
