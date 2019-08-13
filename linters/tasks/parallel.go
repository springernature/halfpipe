package tasks

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintParallelTask(parallelTask manifest.Parallel) (errs []error, warnings []error) {
	for _, task := range parallelTask.Tasks {
		switch task.(type) {
		case manifest.Parallel:
			errs = append(errs, errors.NewInvalidField("type", "You are not allowed to use 'parallel' task inside a 'parallel' task"))
		default:
			if string(task.GetParallelGroup()) != "" {
				warnings = append(warnings, errors.NewInvalidField("parallel", "Please dont use 'parallel' field inside a 'parallel' task!"))
			}
		}
	}

	if len(parallelTask.Tasks) == 0 {
		errs = append(errs, errors.NewInvalidField("tasks", "A 'parallel' task must contain at least one sub task"))
	}

	if len(parallelTask.Tasks) == 1 {
		warnings = append(warnings, errors.NewInvalidField("tasks", "It seems unnecessary to have a single parallel task"))
	}

	return
}
