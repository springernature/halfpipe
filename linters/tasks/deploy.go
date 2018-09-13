package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/manifest"
	"strings"
	"time"
)

func LintDeployCFTask(cf manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
	if cf.API == "" {
		errs = append(errs, errors.NewMissingField("api"))
	}
	if cf.Space == "" {
		errs = append(errs, errors.NewMissingField("space"))
	}
	if cf.Org == "" {
		errs = append(errs, errors.NewMissingField("org"))
	}
	if cf.TestDomain == "" {
		_, found := defaults.DefaultValues.CfTestDomains[cf.API]
		if cf.API != "" && !found {
			errs = append(errs, errors.NewMissingField("testDomain"))
		}
	}

	if cf.Timeout != "" {
		_, err := time.ParseDuration(cf.Timeout)
		if err != nil {
			errs = append(errs, errors.NewInvalidField("timeout", err.Error()))
		}
	}

	if cf.Retries < 0 || cf.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(cf.Manifest, "../artifacts/") {
		warnings = append(warnings, errors.NewFileError(cf.Manifest, "this file must be saved as an artifact in a previous task"))
		if len(cf.PrePromote) > 0 {
			errs = append(errs, errors.NewInvalidField("pre_promote", "if you are using generated manifest you cannot have pre promote tasks"))

		}
	} else {

		if err := filechecker.CheckFile(fs, cf.Manifest, false); err != nil {
			errs = append(errs, err)
		}
	}

	return
}
