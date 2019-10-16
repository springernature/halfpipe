package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetsDefaultDockerComposeService(t *testing.T) {
	assert.Equal(t, DefaultValues.DockerComposeService, dockerComposeDefaulter(manifest.DockerCompose{}, DefaultValues).Service)
}

func TestDoesntOverrideService(t *testing.T) {
	service := "asdf"
	assert.Equal(t, service, dockerComposeDefaulter(manifest.DockerCompose{Service: service}, DefaultValues).Service)
}
