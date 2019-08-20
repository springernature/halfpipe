package triggers

import (
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOnlyCronTriggerDefined(t *testing.T) {
	t.Run("valid trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/10 * * * *"}
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				trigger,
			},
		}

		errs, _ := LintCronTrigger(man, trigger)
		assert.Len(t, errs, 0)
	})

	t.Run("empty trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{}
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				trigger,
			},
		}

		errs, _ := LintCronTrigger(man, trigger)
		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "trigger", errs)
	})

	t.Run("bad trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/99 * * * *"}
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				trigger,
			},
		}

		errs, _ := LintCronTrigger(man, trigger)
		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "trigger", errs)
	})

	t.Run("with seconds in trigger", func(t *testing.T) {
		// 6 parts means there is seconds.
		trigger := manifest.TimerTrigger{Cron: "* * * * * *"}
		man := manifest.Manifest{
			Triggers: manifest.TriggerList{
				trigger,
			},
		}

		errs, _ := LintCronTrigger(man, trigger)
		assert.Len(t, errs, 1)
		helpers.AssertInvalidFieldInErrors(t, "trigger", errs)
	})
}

func TestBothCronTriggerDefined(t *testing.T) {
	// In the merger we dont do anything if both cron_trigger and a TimerTrigger{} is defined.
	// Lets catch that here

	trigger := manifest.TimerTrigger{Cron: "* * * * * *"}
	man := manifest.Manifest{
		CronTrigger: trigger.Cron,
		Triggers: manifest.TriggerList{
			trigger,
		},
	}

	errs, _ := LintCronTrigger(man, trigger)
	assert.Len(t, errs, 1)
	helpers.AssertInvalidFieldInErrors(t, "cron_trigger", errs)
}
