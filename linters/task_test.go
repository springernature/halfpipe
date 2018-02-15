package linters

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

func setup() TaskLinter {
	return TaskLinter{
		Fs: afero.Afero{Fs: afero.NewMemMapFs()},
	}
}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := model.Manifest{}
	taskLinter := setup()

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertMissingField(t, "tasks", result.Errors[0])
}

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := model.Manifest{}
	taskLinter := setup()

	man.Tasks = []model.Task{
		model.Run{},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertMissingField(t, "script", result.Errors[0])
	assertMissingField(t, "image", result.Errors[1])
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Script: "./build.sh",
			Image:  "alpine",
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "./build.sh", result.Errors[0])
}

func TestRunTaskScriptFileExists(t *testing.T) {
	taskLinter := setup()
	taskLinter.Fs.WriteFile("build.sh", []byte("foo"), 0777)

	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Script: "./build.sh",
			Image:  "alpine",
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.DeployCF{},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 3)
	assertMissingField(t, "api", result.Errors[0])
	assertMissingField(t, "space", result.Errors[1])
	assertMissingField(t, "org", result.Errors[2])
}

func TestDockerPushTaskWithEmptyTask(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{
		Tasks: []model.Task{
			model.DockerPush{},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 4)
	assertMissingField(t, "username", result.Errors[0])
	assertMissingField(t, "password", result.Errors[1])
	assertMissingField(t, "repo", result.Errors[2])
	assertFileError(t, "Dockerfile", result.Errors[3])

}

func TestDockerPushTaskWithBadRepo(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{
		Tasks: []model.Task{
			model.DockerPush{
				Username: "asd",
				Password: "asd",
				Repo: "asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 2)
	assertInvalidField(t, "repo", result.Errors[0])
	assertFileError(t, "Dockerfile", result.Errors[1])

}

func TestDockerPushTaskWhenDockerfileIsMissing(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{
		Tasks: []model.Task{
			model.DockerPush{
				Username: "asd",
				Password: "asd",
				Repo: "asd/asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertFileError(t, "Dockerfile", result.Errors[0])
}

func TestDockerPushTaskWithCorrectData(t *testing.T) {
	taskLinter := setup()
	taskLinter.Fs.WriteFile("Dockerfile", []byte("FROM ubuntu"), 0777)

	man := model.Manifest{
		Tasks: []model.Task{
			model.DockerPush{
				Username: "asd",
				Password: "asd",
				Repo: "asd/asd",
			},
		},
	}

	result := taskLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

