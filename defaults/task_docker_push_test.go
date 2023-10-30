package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestWhenPublicImageDontSetUsernameAndPassword(t *testing.T) {
	task := manifest.DockerPush{Image: "asdf", DockerfilePath: "something", ScanTimeout: 15}
	assert.Empty(t, dockerPushDefaulter(task, manifest.Manifest{}, Concourse).Username)
	assert.Empty(t, dockerPushDefaulter(task, manifest.Manifest{}, Concourse).Password)
}

func TestPrivateImageSetsUsernameAndPassword(t *testing.T) {
	task := manifest.DockerPush{Image: path.Join(config.DockerRegistry, "push-me"), DockerfilePath: "something"}
	assert.Equal(t, Concourse.Docker.Username, dockerPushDefaulter(task, manifest.Manifest{}, Concourse).Username)
	assert.Equal(t, Concourse.Docker.Password, dockerPushDefaulter(task, manifest.Manifest{}, Concourse).Password)
}

func TestSetsTheDockerFilePath(t *testing.T) {
	assert.Equal(t, "Dockerfile", dockerPushDefaulter(manifest.DockerPush{}, manifest.Manifest{}, Concourse).DockerfilePath)
}
