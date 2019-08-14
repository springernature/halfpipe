package linters

import (
	"fmt"
	"github.com/mbrevoort/cronexpr"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"reflect"
	"regexp"
)

type cronTriggerLinter struct{}

func NewCronTriggerLinter() cronTriggerLinter {
	return cronTriggerLinter{}
}

func (c cronTriggerLinter) lintOnlyOneCronTrigger(man manifest.Manifest) error {
	numCronTriggers := 0
	var cronTrigger manifest.Trigger

	for _, trigger := range man.Triggers {
		switch trigger.(type) {
		case manifest.Cron:
			cronTrigger = trigger
			numCronTriggers++
		}
	}

	if numCronTriggers > 1 {
		return errors.NewInvalidField("triggers", "You are only allowed one cron trigger")
	}

	if man.CronTrigger != "" && numCronTriggers != 0 && !reflect.DeepEqual(cronTrigger, manifest.Cron{}) {
		return errors.NewInvalidField("cron_trigger/triggers", "You are only allowed to configure cron with either cron_trigger or triggers")
	}

	return nil
}

func (c cronTriggerLinter) getValues(man manifest.Manifest) (CronTrigger, field string) {
	if man.CronTrigger != "" {
		return man.CronTrigger, "cron_trigger"
	}
	// The first thing we do in the linter is to make sure that we use
	// either cron_trigger or manifest.triggers for cron and that there
	// is only one cron trigger, thus we can assume that the first cron
	// trigger we find will be the correct one
	for index, trigger := range man.Triggers {
		switch trigger := trigger.(type) {
		case manifest.Cron:
			return trigger.Trigger, fmt.Sprintf("triggers[%d].trigger", index)
		}
	}
	return
}

func (c cronTriggerLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Cron Trigger Linter"
	result.DocsURL = "https://docs.halfpipe.io/manifest/#cron-trigger"

	if err := c.lintOnlyOneCronTrigger(manifest); err != nil {
		result.AddError(err)
		return
	}

	cronTrigger, field := c.getValues(manifest)

	if cronTrigger != "" {
		_, err := cronexpr.Parse(cronTrigger)
		if err != nil {
			result.AddError(errors.NewInvalidField(field, fmt.Sprintf("%s is not a valid cron expression", cronTrigger)))
		}

		spacer := regexp.MustCompile(`\S+`)
		if len(spacer.FindAllStringIndex(cronTrigger, -1)) == 6 {
			result.AddError(errors.NewInvalidField(field, "seconds in cron expression is not supported"))
		}
	}

	return
}
