package triggers

import (
	"fmt"
	"github.com/mbrevoort/cronexpr"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"regexp"
)

func LintCronTrigger(cron manifest.TimerTrigger) (errs []error, warnings []error) {
	_, err := cronexpr.Parse(cron.Cron)
	if err != nil {
		errs = append(errs, linterrors.NewInvalidField("trigger", fmt.Sprintf("'%s' is not a valid cron expression", cron.Cron)))
	}

	spacer := regexp.MustCompile(`\S+`)
	if len(spacer.FindAllStringIndex(cron.Cron, -1)) == 6 {
		errs = append(errs, linterrors.NewInvalidField("trigger", "seconds in cron expression is not supported"))
	}

	return
}
