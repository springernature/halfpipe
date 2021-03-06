package triggers

import (
	"testing"

	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestOnlyCronTriggerDefined(t *testing.T) {
	t.Run("valid trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/10 * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 0)
	})

	t.Run("empty trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		linterrors.AssertInvalidFieldInErrors(t, "trigger", errs)
	})

	t.Run("bad trigger", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/99 * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		linterrors.AssertInvalidFieldInErrors(t, "trigger", errs)
	})

	t.Run("with seconds in trigger", func(t *testing.T) {
		// 6 parts means there is seconds.
		trigger := manifest.TimerTrigger{Cron: "* * * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		linterrors.AssertInvalidFieldInErrors(t, "trigger", errs)
	})
}
