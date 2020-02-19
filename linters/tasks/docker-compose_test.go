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

func TestDockerCompose_Happy(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml"} //We specify service and compose file here as they are set in the defaulter
	errors, warnings := LintDockerComposeTask(emptyTask, fs, []string{})
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
	errorsAgain, warningsAgain := LintDockerComposeTask(task, fs, []string{})
	assert.Len(t, errorsAgain, 0)
	assert.Len(t, warningsAgain, 0)
}

func TestDockerCompose_MissingFile(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	emptyTask := manifest.DockerCompose{ComposeFile: "docker-compose.yml"}
	errors, _ := LintDockerComposeTask(emptyTask, fs, []string{})
	linterrors.AssertFileErrorInErrors(t, "docker-compose.yml", errors)
}

func TestDockerCompose_UnknownService(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "asdf", ComposeFile: "docker-compose.yml"}
	errors, _ := LintDockerComposeTask(emptyTask, fs, []string{})
	linterrors.AssertInvalidFieldInErrors(t, "service", errors)
}

func TestDockerComposeRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	errors, _ := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: -1}, fs, []string{})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, _ = LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: 6}, fs, []string{})
	linterrors.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, warnings := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml", Retries: 5}, fs, []string{})
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestLintDockerComposeWhenFileIsGarbage(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("foo.yml", []byte("not valid yaml"), 0777)

	errors, _ := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFile: "foo.yml", Retries: 1}, fs, []string{})
	assert.Len(t, errors, 1)
	linterrors.AssertFileErrorInErrors(t, "foo.yml", errors)
}

func TestDockerCompose_DeprecatedDockerRegistry(t *testing.T) {
	composeFile := `
version: 3
services:
  app:
    image: old.registry/curl
  database:
    image: another.old.registry/blah
  ok:
    image: eu.gcr.io/this-is-ok
`
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(composeFile), 0777)

	emptyTask := manifest.DockerCompose{Service: "app", ComposeFile: "docker-compose.yml"} //We specify service and compose file here as they are set in the defaulter
	errors, warnings := LintDockerComposeTask(emptyTask, fs, []string{"old.registry", "another.old.registry"})
	assert.Len(t, errors, 0)
	if assert.Len(t, warnings, 2) {
		assert.Equal(t, linterrors.NewDeprecatedDockerRegistryError("old.registry"), warnings[0])
		assert.Equal(t, linterrors.NewDeprecatedDockerRegistryError("another.old.registry"), warnings[1])
	}
}
