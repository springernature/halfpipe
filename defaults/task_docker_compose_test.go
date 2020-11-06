package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsDefaultDockerComposeService(t *testing.T) {
	assert.Equal(t, DefaultValues.Docker.ComposeService, dockerComposeDefaulter(manifest.DockerCompose{}, DefaultValues).Service)
}

func TestDoesntOverrideService(t *testing.T) {
	service := "asdf"
	assert.Equal(t, service, dockerComposeDefaulter(manifest.DockerCompose{Service: service}, DefaultValues).Service)
}

func TestSetsDefaultDockerComposeFile(t *testing.T) {
	assert.Equal(t, DefaultValues.Docker.ComposeFile, dockerComposeDefaulter(manifest.DockerCompose{}, DefaultValues).ComposeFile)
}

func TestDoesntOverrideDockerComposeFile(t *testing.T) {
	file := "docker-compose-foo.yml"
	assert.Equal(t, file, dockerComposeDefaulter(manifest.DockerCompose{ComposeFile: file}, DefaultValues).ComposeFile)
}
