package linters

import (
	"fmt"
	"github.com/mbrevoort/cronexpr"
	"github.com/springernature/halfpipe/manifest"
	"time"
)

const minCronIntervalMins = 15

func LintCronTrigger(cron manifest.TimerTrigger) (errs []error) {
	expr, err := cronexpr.Parse(cron.Cron)
	if err != nil {
		errs = append(errs, NewErrInvalidField("trigger", fmt.Sprintf("the cron expression '%s' is not valid", cron.Cron)))
		return
	}

	next2 := expr.NextN(time.Now(), 2)
	if next2[1].Before(next2[0].Add(minCronIntervalMins * time.Minute)) {
		errs = append(errs, NewErrInvalidField("trigger", fmt.Sprintf("the cron expression '%s' is more frequent than the minimum interval of 15 minutes", cron.Cron)))
	}
	return errs
}
