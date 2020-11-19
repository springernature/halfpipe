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

		updated := runDefaulter(task, Concourse)
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
				Username: Concourse.Docker.Username,
				Password: Concourse.Docker.Password,
			},
		}

		assert.Equal(t, expectedTask, runDefaulter(task, Concourse))
	})

}
