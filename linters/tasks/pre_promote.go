package tasks

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintPrePromoteTask(task manifest.Task) (errs []error, warnings []error) {
	switch task.(type) {
	case manifest.Run,
		manifest.DockerCompose,
		manifest.ConsumerIntegrationTest:
		if task.IsManualTrigger() {
			errs = append(errs, linterrors.NewInvalidField("manual_trigger", "you are not allowed to have a manual trigger inside a pre promote task"))
		}
		if task.GetNotifications().NotificationsDefined() {
			errs = append(errs, linterrors.NewInvalidField("notifications", "you are not allowed to configure notifications inside a pre promote task"))
		}
	default:
		errs = append(errs, linterrors.NewInvalidField("type", "you are only allowed to use 'run', 'consumer-integration-test' or 'docker-compose' tasks as pre promotes"))
	}

	return errs, warnings
}
