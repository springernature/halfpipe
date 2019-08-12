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
		}
	}
	return
}
