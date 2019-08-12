package tasks

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"testing"
)

func TestLintPrePromoteTasks(t *testing.T) {
	var task manifest.Task
	task = manifest.Run{
		ManualTrigger: true,
	}
	errors, _ := LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "manual_trigger", errors)

	task = manifest.DockerCompose{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "manual_trigger", errors)

	task = manifest.DeployCF{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errors)

	task = manifest.DockerPush{
		ManualTrigger: true,
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errors)

	task = manifest.Parallel{}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errors)
}
