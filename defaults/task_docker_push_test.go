package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestWhenPublicImage(t *testing.T) {
	task := manifest.DockerPush{Image: "asdf", DockerfilePath: "something"}

	assert.Equal(t, task, dockerPushDefaulter(task, DefaultValues))
}

func TestPrivateImage(t *testing.T) {
	task := manifest.DockerPush{Image: path.Join(config.DockerRegistry, "push-me"), DockerfilePath: "something"}

	expected := manifest.DockerPush{
		Image:          path.Join(config.DockerRegistry, "push-me"),
		DockerfilePath: "something",
		Username:       DefaultValues.DockerUsername,
		Password:       DefaultValues.DockerPassword,
	}

	assert.Equal(t, expected, dockerPushDefaulter(task, DefaultValues))
}

func TestSetsTheDockerFilePath(t *testing.T) {
	assert.Equal(t, "Dockerfile", dockerPushDefaulter(manifest.DockerPush{}, DefaultValues).DockerfilePath)
}
