package tasks

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintParallelTask(parallelTask manifest.Parallel) (errs []error, warnings []error) {
	var numSavedArtifacts int
	var numSavedArtifactsOnFailure int
	for _, task := range parallelTask.Tasks {
		switch task.(type) {
		case manifest.Parallel:
			errs = append(errs, linterrors.NewInvalidField("type", "You are not allowed to use 'parallel' task inside a 'parallel' task"))
		default:
			if task.SavesArtifacts() {
				numSavedArtifacts++
			}

			if task.SavesArtifactsOnFailure() {
				numSavedArtifactsOnFailure++
			}
		}
	}

	if len(parallelTask.Tasks) == 0 {
		errs = append(errs, linterrors.NewInvalidField("tasks", "A 'parallel' task must contain at least one sub task"))
	}

	if len(parallelTask.Tasks) == 1 {
		warnings = append(warnings, linterrors.NewInvalidField("tasks", "It seems unnecessary to have a single parallel task"))
	}

	if numSavedArtifacts > 1 {
		warnings = append(warnings, linterrors.NewInvalidField("tasks", "Only one 'parallel' task can save artifacts without ending up in weird race conditions"))
	}

	if numSavedArtifactsOnFailure > 1 {
		warnings = append(warnings, linterrors.NewInvalidField("tasks", "Only one 'parallel' task can save artifacts on failure without ending up in weird race conditions"))
	}

	return
}
