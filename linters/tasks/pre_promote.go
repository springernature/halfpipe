package tasks

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintPrePromoteTask(task manifest.Task) (errs []error, warnings []error) {
	var manualTrigger bool
	switch t := task.(type) {

	case manifest.Run:
		manualTrigger = t.ManualTrigger
	case manifest.DockerCompose:
		manualTrigger = t.ManualTrigger
	case manifest.ConsumerIntegrationTest:
	default:
		errs = append(errs, linterrors.NewInvalidField("type", "You are only allowed to use 'run' or 'docker-compose' tasks as pre promotes"))
	}

	if manualTrigger {
		errs = append(errs, linterrors.NewInvalidField("manual_trigger", "You are not allowed to have a manual trigger inside a pre promote task"))
	}

	return errs, warnings
}
