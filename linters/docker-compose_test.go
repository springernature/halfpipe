package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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

	emptyTask := manifest.DockerCompose{Service: "app", ComposeFiles: []string{"docker-compose.yml"}} //We specify service and compose file here as they are set in the defaulter
	errors := LintDockerComposeTask(emptyTask, fs)
	assert.Len(t, errors, 0)

	task := manifest.DockerCompose{
		Name:         "run docker compose",
		Service:      "app",
		ComposeFiles: []string{"docker-compose.yml"},
		Vars: manifest.Vars{
			"A": "a",
			"B": "b",
		},
	}
	errors = LintDockerComposeTask(task, fs)
	assert.Len(t, errors, 0)

}

func TestDockerCompose_MissingFile(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{ComposeFiles: []string{"missing.yml", "docker-compose.yml", "missing2.yml"}}
	errors := LintDockerComposeTask(emptyTask, fs)
	assertContainsError(t, errors, ErrFileNotFound.WithFile("missing.yml"))
	assertContainsError(t, errors, ErrFileNotFound.WithFile("missing2.yml"))
}

func TestDockerCompose_UnknownService(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)
	fs.WriteFile("missing-service.yml", []byte(strings.ReplaceAll(validDockerCompose, "app:", "xxx:")), 0777)

	emptyTask := manifest.DockerCompose{Service: "asdf", ComposeFiles: []string{"missing-service.yml", "docker-compose.yml", "missing-service.yml"}}
	errors := LintDockerComposeTask(emptyTask, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("service"))
}

func TestDockerComposeRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	errors := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFiles: []string{"docker-compose.yml"}, Retries: -1}, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	errors = LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFiles: []string{"docker-compose.yml"}, Retries: 6}, fs)
	assertContainsError(t, errors, ErrInvalidField.WithValue("retries"))

	errors = LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFiles: []string{"docker-compose.yml"}, Retries: 5}, fs)
	assert.Len(t, errors, 0)
}

func TestDockerComposeWithoutServicesKey(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(invalidDockerCompose), 0777)

	errors := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFiles: []string{"docker-compose.yml"}}, fs)
	assertContainsError(t, errors, ErrDockerComposeVersion)
}

func TestLintDockerComposeWhenFileIsGarbage(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("foo.yml", []byte("not valid yaml"), 0777)

	errors := LintDockerComposeTask(manifest.DockerCompose{Service: "app", ComposeFiles: []string{"foo.yml"}, Retries: 1}, fs)
	assertContainsError(t, errors, ErrFileInvalid)
}
