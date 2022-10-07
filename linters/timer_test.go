package linters

import (
	"testing"

	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestCronTrigger(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/30 * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 0)
	})

	t.Run("empty", func(t *testing.T) {
		trigger := manifest.TimerTrigger{}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		AssertContainsError(t, errs, ErrInvalidField.WithValue("trigger"))
	})

	t.Run("bad", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/99 * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		AssertContainsError(t, errs, ErrInvalidField.WithValue("trigger"))
	})

	t.Run("too frequent", func(t *testing.T) {
		trigger := manifest.TimerTrigger{Cron: "*/10 * * * *"}

		errs, _ := LintCronTrigger(trigger)
		assert.Len(t, errs, 1)
		AssertContainsError(t, errs, ErrInvalidField.WithValue("trigger"))
	})
}
