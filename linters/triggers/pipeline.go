package triggers

import (
	"fmt"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

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
	return
}
