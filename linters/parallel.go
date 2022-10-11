package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintParallelTask(parallelTask manifest.Parallel) (errs []error) {
	var numSavedArtifacts int
	var numSavedArtifactsOnFailure int
	for _, task := range parallelTask.Tasks {
		switch task.(type) {
		case manifest.Parallel:
			errs = append(errs, NewErrInvalidField("type", "you are not allowed to use 'parallel' task inside a 'parallel' task"))
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
		errs = append(errs, NewErrInvalidField("tasks", "a 'parallel' task must contain at least one sub task"))
	}

	if len(parallelTask.Tasks) == 1 {
		errs = append(errs, NewErrInvalidField("tasks", "it seems unnecessary to have a single parallel task").AsWarning())
	}

	if numSavedArtifacts > 1 {
		errs = append(errs, NewErrInvalidField("tasks", "only one 'parallel' task can save artifacts without ending up in weird race conditions").AsWarning())
	}

	if numSavedArtifactsOnFailure > 1 {
		errs = append(errs, NewErrInvalidField("tasks", "only one 'parallel' task can save artifacts on failure without ending up in weird race conditions").AsWarning())
	}

	return errs
}
