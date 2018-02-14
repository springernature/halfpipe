package linters

import (
	"testing"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
)

var taskLinter = TaskLinter{}

func TestAtLeastOneTaskExists(t *testing.T) {
	man := model.Manifest{}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 1)
	assert.IsType(t, errors.MissingField{}, errs[0])
}

func TestRunTaskWithoutScriptAndImage(t *testing.T) {
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Name:   "",
			Script: "",
			Image:  "",
			Vars:   nil,
		},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 2)

	assert.IsType(t, errors.MissingField{}, errs[0])
	assert.IsType(t, errors.MissingField{}, errs[1])
}

func TestRunTaskWithScriptAndImage(t *testing.T) {
	man := model.Manifest{}
	man.Tasks = []model.Task{
		model.Run{
			Name:   "",
			Script: "./build.sh",
			Image:  "alpine",
			Vars:   nil,
		},
	}

	errs := taskLinter.Lint(man)
	assert.Len(t, errs, 0)
}
