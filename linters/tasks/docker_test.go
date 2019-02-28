package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, _ := LintDockerPushTask(manifest.DockerPush{}, fs)

	helpers.AssertMissingFieldInErrors(t, "username", errors)
	helpers.AssertMissingFieldInErrors(t, "password", errors)
	helpers.AssertMissingFieldInErrors(t, "image", errors)
}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd",
	}

	errors, _ := LintDockerPushTask(task, fs)
	helpers.AssertInvalidFieldInErrors(t, "image", errors)
}

func TestDockerPushTaskWhenDockerfileIsMissing(t *testing.T) {
	t.Run("When DockerfilePath is just Dockerfile", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, fs)

		helpers.AssertFileErrorInErrors(t, "Dockerfile", errors)
	})

	t.Run("When DockerfilePath is in a different folder", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "dockerfiles/Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, fs)

		helpers.AssertFileErrorInErrors(t, "dockerfiles/Dockerfile", errors)
	})

	t.Run("When DockerfilePath is in a different folder upwards", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}

		task := manifest.DockerPush{
			Username:       "asd",
			Password:       "asd",
			Image:          "user/image",
			DockerfilePath: "../dockerfiles/Dockerfile",
		}

		errors, _ := LintDockerPushTask(task, fs)

		helpers.AssertFileErrorInErrors(t, "../dockerfiles/Dockerfile", errors)
	})
}

func TestDockerPushTaskWithCorrectData(t *testing.T) {

	t.Run("When DockerfilePath is just Dockerfile", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "Dockerfile",
		}

		errors, warnings := LintDockerPushTask(task, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("When DockerfilePath is in a different path", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("dockerfile/Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "dockerfile/Dockerfile",
		}

		errors, warnings := LintDockerPushTask(task, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

	t.Run("When DockerfilePath is in a different folder upwards", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile("../dockerfile/Dockerfile", []byte("FROM ubuntu"), 0777)

		task := manifest.DockerPush{
			Username: "asd",
			Password: "asd",
			Image:    "asd/asd",
			Vars: map[string]string{
				"A": "a",
				"B": "b",
			},
			DockerfilePath: "../dockerfile/Dockerfile",
		}

		errors, warnings := LintDockerPushTask(task, fs)
		assert.Len(t, errors, 0)
		assert.Len(t, warnings, 0)
	})

}

func TestDockerPushRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	task := manifest.DockerPush{
		Username: "asd",
		Password: "asd",
		Image:    "asd/asd",
		Vars: map[string]string{
			"A": "a",
			"B": "b",
		},
		DockerfilePath: "Dockerfile",
	}

	task.Retries = -1
	errors, _ := LintDockerPushTask(task, fs)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 6
	errors, _ = LintDockerPushTask(task, fs)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	task.Retries = 4
	errors, warnings := LintDockerPushTask(task, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

var validDockerCompose = `
version: 3
services:
  app:
    image: appropriate/curl`

func TestDockerCompose_Happy(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "app"} //We specify service here as its default is set in the defaulter
	errors, warnings := LintDockerComposeTask(emptyTask, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)

	task := manifest.DockerCompose{
		Name:    "run docker compose",
		Service: "app",
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

	emptyTask := manifest.DockerCompose{}
	errors, _ := LintDockerComposeTask(emptyTask, fs)
	helpers.AssertFileErrorInErrors(t, "docker-compose.yml", errors)
}

func TestDockerCompose_UnknownService(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	emptyTask := manifest.DockerCompose{Service: "asdf"}
	errors, _ := LintDockerComposeTask(emptyTask, fs)
	helpers.AssertInvalidFieldInErrors(t, "service", errors)
}

func TestDockerComposeRetries(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("docker-compose.yml", []byte(validDockerCompose), 0777)

	errors, _ := LintDockerComposeTask(manifest.DockerCompose{Service: "app", Retries: -1}, fs)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, _ = LintDockerComposeTask(manifest.DockerCompose{Service: "app", Retries: 6}, fs)
	helpers.AssertInvalidFieldInErrors(t, "retries", errors)

	errors, warnings := LintDockerComposeTask(manifest.DockerCompose{Service: "app", Retries: 5}, fs)
	assert.Len(t, errors, 0)
	assert.Len(t, warnings, 0)
}

func TestLintDockerComposeServiceWhenFileError(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	errors, _ := lintDockerComposeService("someService", "../bar.yml", fs)
	assert.Len(t, errors, 1)
	assert.Equal(t, "open ../bar.yml: file does not exist", errors[0].Error())
}

func TestLintDockerComposeServiceWhenFileIsGarbage(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile("foo.yml", []byte("not valid yaml"), 0777)

	errors, _ := lintDockerComposeService("someService", "foo.yml", fs)
	assert.Len(t, errors, 1)
	helpers.AssertFileErrorInErrors(t, "foo.yml", errors)
}
