package tasks

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintPrePromoteTask(task manifest.Task) (errs []error, warnings []error) {
	var manual_trigger bool
	var parallel bool
	switch t := task.(type) {

	case manifest.Run:
		manual_trigger = t.ManualTrigger
		parallel = t.Parallel
	case manifest.DockerCompose:
		manual_trigger = t.ManualTrigger
		parallel = t.Parallel
	case manifest.ConsumerIntegrationTest:
		parallel = t.Parallel
	default:
		errs = append(errs, errors.NewInvalidField("type", "You are only allowed to use 'run' or 'docker-compose' tasks as pre promotes"))
	}

	if manual_trigger {
		errs = append(errs, errors.NewInvalidField("manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
	}
	if parallel {
		errs = append(errs, errors.NewInvalidField("parallel", "You are not allowed to set this inside a pre promote task"))
	}

	return
}
