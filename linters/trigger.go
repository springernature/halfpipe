package linters

import (
	"fmt"
	"github.com/mbrevoort/cronexpr"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"regexp"
)

type triggerLinter struct{}

func NewTriggerLinter() triggerLinter {
	return triggerLinter{}
}

func (triggerLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Timer Trigger Linter"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#trigger-interval-deprecated , https://docs.halfpipe.io/manifest/#cron-trigger"

	if manifest.TriggerInterval != "" && manifest.CronTrigger != "" {
		result.AddError(errors.NewInvalidField("trigger_interval", "please remove trigger_interval if you use cron_trigger"))
	} else if manifest.TriggerInterval != "" {
		result.AddWarning(errors.NewInvalidField("trigger_interval", "this field is deprecated, please use 'cron_trigger' instead"))
	}

	if manifest.CronTrigger != "" {
		_, err := cronexpr.Parse(manifest.CronTrigger)
		if err != nil {
			result.AddError(errors.NewInvalidField("cron_trigger", fmt.Sprintf("%s is not a valid cron expression", manifest.CronTrigger)))
		}

		spacer := regexp.MustCompile(`\S+`)
		if len(spacer.FindAllStringIndex(manifest.CronTrigger, -1)) == 6 {
			result.AddError(errors.NewInvalidField("cron_trigger", "seconds in cron expression is not supported"))
		}
	}

	return
}
