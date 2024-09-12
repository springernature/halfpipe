package linters

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"slices"
	"strings"
)

func LintPipelineTrigger(man manifest.Manifest, pipeline manifest.PipelineTrigger) (errs []error) {
	if man.Team != pipeline.Team {
		errs = append(errs, NewErrInvalidField("team", fmt.Sprintf("you can only trigger on pipelines in your team, '%s'!", man.Team)))
		return errs
	}

	if pipeline.Pipeline == "" {
		errs = append(errs, NewErrInvalidField("pipeline", "must not be empty"))
		return errs
	}

	if pipeline.Job == "" {
		errs = append(errs, NewErrInvalidField("job", "must not be empty"))
		return errs
	}

	allowedStatus := []string{"succeeded", "failed", "errored", "aborted"}
	if !slices.Contains(allowedStatus, pipeline.Status) {
		errs = append(errs, NewErrInvalidField("status", fmt.Sprintf("must be one of %s", strings.Join(allowedStatus, ", "))))
	}
	return errs
}
