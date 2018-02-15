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

func (taskLinter TaskLinter) Lint(man model.Manifest) (result errors.LintResult) {
	result.Linter = "Tasks Linter"

	if len(man.Tasks) == 0 {
		result.Errors = append(result.Errors, errors.NewMissingField("tasks"))
		return
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.Run:
			result.Errors = append(result.Errors, lintRunTask(taskLinter, task)...)
		case model.DeployCF:
			result.Errors = append(result.Errors, lintDeployCFTask(task)...)
		default:
			result.Errors = append(result.Errors, errors.NewInvalidField("task", fmt.Sprintf("task %v '%s' is not a known task", i+1, task.GetName())))
		}
	}

	return
}
func lintDeployCFTask(cf model.DeployCF) (errs []error) {
	if cf.Api == "" {
		errs = append(errs, errors.NewMissingField("api"))
	}

	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("org"))
	}
	return
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
