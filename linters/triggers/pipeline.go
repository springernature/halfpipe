package triggers

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/errors"
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
		errs = append(errs, errors.NewInvalidField("team", fmt.Sprintf("you can only trigger on pipelines in your team, '%s'!", man.Team)))
		return
	}

	if pipeline.Pipeline == "" {
		errs = append(errs, errors.NewInvalidField("pipeline", "must not be empty"))
		return
	}

	if pipeline.Job == "" {
		errs = append(errs, errors.NewInvalidField("job", "must not be empty"))
		return
	}

	allowedStatus := []string{"succeeded", "failed", "errored", "aborted"}
	if !contains(allowedStatus, pipeline.Status) {
		errs = append(errs, errors.NewInvalidField("status", fmt.Sprintf("must be one of %s", strings.Join(allowedStatus, ", "))))
	}
	return
}
