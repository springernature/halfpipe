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

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assertMissingField(t, "tasks", errs[0])
}

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := model.Manifest{}
	taskLinter := setup()

	man.Tasks = []model.Task{
		model.Run{},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 2)
	assertMissingField(t, "script", errs[0])
	assertMissingField(t, "image", errs[1])
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

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assertFileError(t, "./build.sh", errs[0])
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

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 0)
}

func TestCFDeployTaskWithEmptyTask(t *testing.T) {
	taskLinter := setup()
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.DeployCF{},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 3)
	assertMissingField(t, "api", errs[0])
	assertMissingField(t, "space", errs[1])
	assertMissingField(t, "org", errs[2])
}
