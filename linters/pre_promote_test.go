package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
)

func TestLintPrePromoteTasks(t *testing.T) {

	t.Run("Manual trigger", func(t *testing.T) {
		task := manifest.Run{
			ManualTrigger: true,
		}
		errors := LintPrePromoteTask(task)
		assertContainsError(t, errors, ErrInvalidField.WithValue("manual_trigger"))
	})

	t.Run("Notifications", func(t *testing.T) {
		task := manifest.Run{
			Notifications: manifest.Notifications{
				Slack: manifest.Slack{
					OnSuccess: []string{
						"#kehe",
					},
				},
			},
		}
		errors := LintPrePromoteTask(task)
		assertContainsError(t, errors, ErrInvalidField.WithValue("notifications"))
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
			errors := LintPrePromoteTask(nonSupportedTask)
			assertContainsError(t, errors, ErrInvalidField.WithValue("type"))
		}

	})

}
