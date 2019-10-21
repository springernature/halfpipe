package triggers

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func contains(allowedStatus []string, status string) bool {
	for _, a := range allowedStatus {
		if a == status {
			return true
		}
	}
	return false
}

func LintPipelineTrigger(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error, warnings []error) {
	if man.Team != pipeline.Team {
		errs = append(errs, linterrors.NewInvalidField("team", fmt.Sprintf("you can only trigger on pipelines in your team, '%s'!", man.Team)))
		return errs, warnings
	}

	if pipeline.Pipeline == "" {
		errs = append(errs, linterrors.NewInvalidField("pipeline", "must not be empty"))
		return errs, warnings
	}

	if pipeline.Job == "" {
		errs = append(errs, linterrors.NewInvalidField("job", "must not be empty"))
		return errs, warnings
	}

	allowedStatus := []string{"succeeded", "failed", "errored", "aborted"}
	if !contains(allowedStatus, pipeline.Status) {
		errs = append(errs, linterrors.NewInvalidField("status", fmt.Sprintf("must be one of %s", strings.Join(allowedStatus, ", "))))
	}
	return errs, warnings
}
