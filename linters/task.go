package linters

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type TaskLinter struct {
	Fs afero.Afero
}

func (taskLinter TaskLinter) Lint(man model.Manifest) []error {
	var errs []error
	if len(man.Tasks) == 0 {
		errs = append(errs, errors.NewMissingField("tasks"))
		return errs
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.Run:
			errs = append(errs, lintRunTask(taskLinter, task)...)
		case model.DeployCF:
			errs = append(errs, lintDeployCFTask(task)...)
		default:
			errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("task %v '%s' is not a known task", i+1, task.GetName())))
		}
	}

	return errs
}
func lintDeployCFTask(cf model.DeployCF) []error {
	var errs []error
	if cf.Api == "" {
		errs = append(errs, errors.NewMissingField("api"))
	}

	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("org"))
	}
	return errs
}

func lintRunTask(t TaskLinter, run model.Run) []error {
	var errs []error
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	} else {
		if err := CheckFile(t.Fs, run.Script, true); err != nil {
			errs = append(errs, err)
		}
	}

	if run.Image == "" {
		errs = append(errs, errors.NewMissingField("image"))
	}

	return errs
}
