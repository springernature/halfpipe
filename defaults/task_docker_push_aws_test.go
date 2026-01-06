package defaults

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDockerPushAWSDefaultsAWSFieldsWhenEmpty(t *testing.T) {
	task := manifest.DockerPushAWS{}
	man := manifest.Manifest{}

	result := dockerPushAWSDefaulter(task, man, Actions)

	assert.Equal(t, Actions.AWSDocker.Region, result.Region)
	assert.Equal(t, Actions.AWSDocker.AccessKeyID, result.AccessKeyID)
	assert.Equal(t, Actions.AWSDocker.SecretAccessKey, result.SecretAccessKey)
	assert.Equal(t, "Dockerfile", result.DockerfilePath)
}

func TestDockerPushAWSPreservesExistingValues(t *testing.T) {
	task := manifest.DockerPushAWS{
		Region:          "eu-west-1",
		AccessKeyID:     "custom-key-id",
		SecretAccessKey: "custom-secret",
		DockerfilePath:  "SomePath",
	}
	man := manifest.Manifest{}

	result := dockerPushAWSDefaulter(task, man, Actions)

	assert.Equal(t, "eu-west-1", result.Region)
	assert.Equal(t, "custom-key-id", result.AccessKeyID)
	assert.Equal(t, "custom-secret", result.SecretAccessKey)
	assert.Equal(t, "SomePath", result.DockerfilePath)
}
