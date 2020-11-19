package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsDefaultDockerComposeService(t *testing.T) {
	assert.Equal(t, Concourse.Docker.ComposeService, dockerComposeDefaulter(manifest.DockerCompose{}, Concourse).Service)
}

func TestDoesntOverrideService(t *testing.T) {
	service := "asdf"
	assert.Equal(t, service, dockerComposeDefaulter(manifest.DockerCompose{Service: service}, Concourse).Service)
}

func TestSetsDefaultDockerComposeFile(t *testing.T) {
	assert.Equal(t, Concourse.Docker.ComposeFile, dockerComposeDefaulter(manifest.DockerCompose{}, Concourse).ComposeFile)
}

func TestDoesntOverrideDockerComposeFile(t *testing.T) {
	file := "docker-compose-foo.yml"
	assert.Equal(t, file, dockerComposeDefaulter(manifest.DockerCompose{ComposeFile: file}, Concourse).ComposeFile)
}
