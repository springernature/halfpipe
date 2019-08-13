package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCronTrigger(t *testing.T) {
	t.Run("cron_trigger", func(t *testing.T) {
		man := manifest.Manifest{}
		man.CronTrigger = "*/10 * * * *"

		assert.False(t, NewCronTriggerLinter().Lint(man).HasErrors())
	})

	t.Run("triggers", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.Cron{
					Trigger: "*/10 * * * *",
				},
			},
		}

		assert.False(t, NewCronTriggerLinter().Lint(man).HasErrors())
	})
}

func TestInvalidCronTrigger(t *testing.T) {
	t.Run("cron_trigger", func(t *testing.T) {
		man := manifest.Manifest{}
		man.CronTrigger = "*/99 * * * *"

		result := NewCronTriggerLinter().Lint(man)
		assertInvalidFieldInErrors(t, "cron_trigger", result.Errors)
	})

	t.Run("triggers", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.Git{},
				manifest.Cron{
					Trigger: "*/99 * * * *",
				},
			},
		}
		result := NewCronTriggerLinter().Lint(man)
		assertInvalidFieldInErrors(t, "triggers[1].trigger", result.Errors)

	})
}

func TestCronTriggerWithSecondsShouldHaveError(t *testing.T) {
	t.Run("cron_trigger", func(t *testing.T) {
		man := manifest.Manifest{}
		// 6 parts means there is seconds.
		man.CronTrigger = "* * * * * *"
		result := NewCronTriggerLinter().Lint(man)
		assertInvalidFieldInErrors(t, "cron_trigger", result.Errors)
	})

	t.Run("triggers", func(t *testing.T) {
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				manifest.Cron{
					Trigger: "* * * * * *",
				},
			},
		}

		result := NewCronTriggerLinter().Lint(man)
		assertInvalidFieldInErrors(t, "triggers[0].trigger", result.Errors)
	})
}

func TestOnlyAllowedOneCronTrigger(t *testing.T) {
	man := manifest.Manifest{
		Triggers: manifest.TriggerList{
			manifest.Cron{},
			manifest.Cron{},
		},
	}

	result := NewCronTriggerLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidFieldInErrors(t, "triggers", result.Errors)
}

func TestOnlyAllowsEitherCronTriggerOrCron(t *testing.T) {
	man := manifest.Manifest{
		CronTrigger: "asdf",
		Triggers: manifest.TriggerList{
			manifest.Cron{
				Trigger: "asdf",
			},
		},
	}

	result := NewCronTriggerLinter().Lint(man)
	assert.Len(t, result.Errors, 1)
	assertInvalidFieldInErrors(t, "cron_trigger/triggers", result.Errors)
}
