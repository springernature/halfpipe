package triggers

import (
	"fmt"
	"github.com/mbrevoort/cronexpr"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
	"regexp"
)

//func getValues(man manifest.Manifest) (CronTrigger) {
//	if man.CronTrigger != "" {
//		return man.CronTrigger
//	}
//	// The first thing we do in the linter is to make sure that we use
//	// either cron_trigger or manifest.triggers for cron and that there
//	// is only one cron trigger, thus we can assume that the first cron
//	// trigger we find will be the correct one
//	for index, trigger := range man.Triggers {
//		switch trigger := trigger.(type) {
//		case manifest.CronTrigger:
//			return trigger.Trigger
//		}
//	}
//	return
//}

func LintCronTrigger(man manifest.Manifest, cron manifest.CronTrigger) (errs []error, warnings []error) {
	/*
		in the trigger translator we do the following
		only cron_trigger: x defined -> CronTrigger{x}
		cron_trigger: x defined and CronTrigger{y} -> cron_trigger:x, CronTrigger{y}
		only CronTrigger{y} defined  -> CronTrigger{y}
	*/
	if man.CronTrigger != "" {
		errs = append(errs, errors.NewInvalidField("cron_trigger", "looks like both top level field 'cron_trigger' and a cron trigger is defined. Please remove 'cron_trigger'!"))
		return
	}

	_, err := cronexpr.Parse(cron.Trigger)
	if err != nil {
		errs = append(errs, errors.NewInvalidField("trigger", fmt.Sprintf("'%s' is not a valid cron expression", cron.Trigger)))
	}

	spacer := regexp.MustCompile(`\S+`)
	if len(spacer.FindAllStringIndex(cron.Trigger, -1)) == 6 {
		errs = append(errs, errors.NewInvalidField("trigger", "seconds in cron expression is not supported"))
	}

	//if cronTrigger != "" {
	//	_, err := cronexpr.Parse(cronTrigger)
	//	if err != nil {
	//		errs = append(errs, errors.NewInvalidField(field, fmt.Sprintf("%s is not a valid cron expression", cronTrigger)))
	//	}
	//
	//	spacer := regexp.MustCompile(`\S+`)
	//	if len(spacer.FindAllStringIndex(cronTrigger, -1)) == 6 {
	//		errs = append(errs, errors.NewInvalidField(field, "seconds in cron expression is not supported"))
	//	}
	//}

	return
}
