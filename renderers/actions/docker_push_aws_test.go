package actions

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestDockerPushAWSSteps_BuildArgsFromVars(t *testing.T) {
	a := Actions{workingDir: "."}
	task := manifest.DockerPushAWS{
		Image: "my-repo",
		Vars: manifest.Vars{
			"MY_VAR":      "my-value",
			"ANOTHER_VAR": "another-value",
		},
	}
	man := manifest.Manifest{Team: "test-team"}

	steps := a.dockerPushAWSSteps(task, man)

	var buildStep Step
	for _, s := range steps {
		if s.Name == "Build, tag, and push Docker image" {
			buildStep = s
			break
		}
	}

	assert.NotEmpty(t, buildStep.Name, "should have a build step")

	assert.Contains(t, buildStep.Run, "--build-arg MY_VAR=my-value")
	assert.Contains(t, buildStep.Run, "--build-arg ANOTHER_VAR=another-value")

	assert.Contains(t, buildStep.Run, "--build-arg ARTIFACTORY_PASSWORD")
	assert.Contains(t, buildStep.Run, "--build-arg ARTIFACTORY_URL")
	assert.Contains(t, buildStep.Run, "--build-arg ARTIFACTORY_USERNAME")
	assert.Contains(t, buildStep.Run, "--build-arg BUILD_VERSION")
	assert.Contains(t, buildStep.Run, "--build-arg GIT_REVISION")
	assert.Contains(t, buildStep.Run, "--build-arg RUNNING_IN_CI")
}

func TestDockerPushAWSSteps_SecretsAsDockerSecrets(t *testing.T) {
	a := Actions{workingDir: "."}
	task := manifest.DockerPushAWS{
		Image: "my-repo",
		Secrets: manifest.Vars{
			"SECRET_A": "secret-value-a",
			"SECRET_B": "secret-value-b",
		},
	}
	man := manifest.Manifest{Team: "test-team"}

	steps := a.dockerPushAWSSteps(task, man)

	var buildStep Step
	for _, s := range steps {
		if s.Name == "Build, tag, and push Docker image" {
			buildStep = s
			break
		}
	}

	assert.NotEmpty(t, buildStep.Name, "should have a build step")

	assert.Contains(t, buildStep.Run, "--secret id=SECRET_A,env=SECRET_A")
	assert.Contains(t, buildStep.Run, "--secret id=SECRET_B,env=SECRET_B")

	assert.Equal(t, "secret-value-a", buildStep.Env["SECRET_A"])
	assert.Equal(t, "secret-value-b", buildStep.Env["SECRET_B"])
}

func TestDockerPushAWSSteps_VarsAndSecretsTogether(t *testing.T) {
	a := Actions{workingDir: "."}
	task := manifest.DockerPushAWS{
		Image: "my-repo",
		Vars: manifest.Vars{
			"MY_VAR": "my-value",
		},
		Secrets: manifest.Vars{
			"MY_SECRET": "secret-value",
		},
	}
	man := manifest.Manifest{Team: "test-team"}

	steps := a.dockerPushAWSSteps(task, man)

	var buildStep Step
	for _, s := range steps {
		if s.Name == "Build, tag, and push Docker image" {
			buildStep = s
			break
		}
	}

	assert.NotEmpty(t, buildStep.Name, "should have a build step")

	assert.Contains(t, buildStep.Run, "--build-arg MY_VAR=my-value")
	assert.Contains(t, buildStep.Run, "--secret id=MY_SECRET,env=MY_SECRET")
	assert.Equal(t, "secret-value", buildStep.Env["MY_SECRET"])
}

func TestDockerPushAWSSteps_NoVarsOrSecrets(t *testing.T) {
	a := Actions{workingDir: "."}
	task := manifest.DockerPushAWS{
		Image: "my-repo",
	}
	man := manifest.Manifest{Team: "test-team"}

	steps := a.dockerPushAWSSteps(task, man)

	var buildStep Step
	for _, s := range steps {
		if s.Name == "Build, tag, and push Docker image" {
			buildStep = s
			break
		}
	}

	assert.NotEmpty(t, buildStep.Name, "should have a build step")

	assert.Contains(t, buildStep.Run, "--build-arg ARTIFACTORY_PASSWORD")

	assert.NotContains(t, buildStep.Run, "--secret id=")
}
