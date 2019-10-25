package tasks

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func TestLintPrePromoteTasks(t *testing.T) {

	t.Run("Manual trigger", func(t *testing.T) {
		task := manifest.Run{
			ManualTrigger: true,
		}
		errors, _ := LintPrePromoteTask(task)
		linterrors.AssertInvalidFieldInErrors(t, "manual_trigger", errors)
	})

	t.Run("Notifications", func(t *testing.T) {
		task := manifest.Run{
			Notifications: manifest.Notifications{
				OnSuccess: []string{
					"#kehe",
				},
			},
		}
		errors, _ := LintPrePromoteTask(task)
		linterrors.AssertInvalidFieldInErrors(t, "notifications", errors)
	})

	t.Run("Non supported task", func(t *testing.T) {
		nonSupportedTasks := manifest.TaskList{
			manifest.DeployCF{},
			manifest.DockerPush{},
			manifest.DeployMLZip{},
			manifest.DeployMLModules{},
			manifest.Parallel{},
			manifest.Sequence{},
		}

		for _, nonSupportedTask := range nonSupportedTasks {
			errors, _ := LintPrePromoteTask(nonSupportedTask)
			linterrors.AssertInvalidFieldInErrors(t, "type", errors)
		}

	})

}
