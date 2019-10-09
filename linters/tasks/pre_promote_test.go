package tasks

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func TestLintPrePromoteTasks(t *testing.T) {
	var task manifest.Task
	task = manifest.Run{
		ManualTrigger: true,
	}
	errors, _ := LintPrePromoteTask(task)
	linterrors.AssertInvalidFieldInErrors(t, "manual_trigger", errors)

	task = manifest.DockerCompose{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	linterrors.AssertInvalidFieldInErrors(t, "manual_trigger", errors)

	task = manifest.DeployCF{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	linterrors.AssertInvalidFieldInErrors(t, "type", errors)

	task = manifest.DockerPush{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	linterrors.AssertInvalidFieldInErrors(t, "type", errors)

	task = manifest.Parallel{}
	errors, _ = LintPrePromoteTask(task)
	linterrors.AssertInvalidFieldInErrors(t, "type", errors)
}
