package linters

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
)

type triggerLinter struct{}

func NewTriggerLinter() triggerLinter {
	return triggerLinter{}
}

func (triggerLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	if manifest.TriggerInterval != "" && manifest.CronTrigger != "" {
		result.AddError(errors.NewInvalidField("trigger_interval", "please remove trigger_interval if you use cron_trigger"))
	} else if manifest.TriggerInterval != "" {
		result.AddWarning(errors.NewInvalidField("trigger_interval", "trigger_interval is deprecated, please use cron_trigger instead"))
	}

	return
}
