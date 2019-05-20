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
		Parallel:      "true",
	}
	errors, _ := LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "manual_trigger", errors)
	helpers.AssertInvalidFieldInErrors(t, "parallel", errors)

	task = manifest.DockerCompose{
		ManualTrigger: true,
		Parallel:      "true",
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "manual_trigger", errors)
	helpers.AssertInvalidFieldInErrors(t, "parallel", errors)

	task = manifest.DeployCF{
		ManualTrigger: true,
		Parallel:      "true",
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errors)

	task = manifest.DockerPush{
		ManualTrigger: true,
		Parallel:      "true",
	}
	errors, _ = LintPrePromoteTask(task)
	helpers.AssertInvalidFieldInErrors(t, "type", errors)
}
