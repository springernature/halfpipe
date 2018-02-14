package linters

import (
	"fmt"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
)

type TaskLinter struct{}

func (t TaskLinter) Lint(man model.Manifest) []error {
	var errs []error
	if len(man.Tasks) == 0 {
		errs = append(errs, errors.NewMissingField("tasks"))
		return errs
	}

	for i, t := range man.Tasks {
		switch task := t.(type) {
		case model.Run:
			errs = append(errs, lintRunTask(task)...)
		default:
			errs = append(errs, errors.NewInvalidField("task", fmt.Sprintf("task %v '%s' is not a known task", i+1, task.GetName())))
		}
	}
	//loop through tasks
	//lint them individually

	return errs
}

func lintRunTask(run model.Run) []error {
	var errs []error
	if run.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	}
	if run.Image == "" {
		errs = append(errs, errors.NewMissingField("image"))
	}
	return errs
}
