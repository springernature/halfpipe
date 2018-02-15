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
