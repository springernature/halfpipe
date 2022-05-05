package tasks

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

var validDockerCompose = `
version: 3
services:
  app:
    image: appropriate/curl`

var invalidDockerCompose = `
app:
  image: appropriate/curl`

func TestDockerCompose_Happy(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml"} //We specify service and compose file here as they are set in the defaulter
	errors, warnings := LintDockerComposeTask(emptyTask, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)

	task := manifest.DockerCompose{
		Name:        "run docker compose",
		Service:     "app",
		ComposeFile: "docker-compose.yml",
		Vars: manifest.Vars{
			"A": "a",
			"B": "b",
		},
	}
	errorsAgain, warningsAgain := LintDockerComposeTask(task, fs)
	assert.Len(t, errorsAgain, 0)
	assert.Len(t, warningsAgain, 0)
}

func TestDockerCompose_MissingFile(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	emptyTask := manifest.DockerCompose{ComposeFile: "docker-compose.yml"}
	errors, _ := LintDockerComposeTask(emptyTask, fs)
	linterrors.AssertFileErrorInErrors(t, "docker-compose.yml", errors)
}

func TestDockerCompose_UnknownService(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "asdf", ComposeFile: "docker-compose.yml"}
	errors, _ := LintDockerComposeTask(emptyTask, fs)
	linterrors.AssertInvalidFieldInErrors(t, "service", errors)
}

func TestDockerComposeRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	errors, _ := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: -1}, fs)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, _ = LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: 6}, fs)
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, warnings := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: 5}, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestDockerComposeWithoutServicesKey(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(invalidDockerCompose), 0777)

	_, warnings := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml"}, fs)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings, linterrors.DeprecatedDockerComposeVersionError{})
}

func TestLintDockerComposeWhenFileIsGarbage(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("foo.yml", []byte("not valid yaml"), 0777)

	errors, _ := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "foo.yml", Retries: 1}, fs)
	assert.Len(t, errors, 1)
	linterrors.AssertFileErrorInErrors(t, "foo.yml", errors)
}
