package tasks

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func LintDeployCFTask(cf manifest.DeployCF, fs afero.Afero) (errs []error, warnings []error) {
	if cf.API == "" {
		errs = append(errs, linterrors.NewMissingField("api"))
	}
	if cf.Space == "" {
		errs = append(errs, linterrors.NewMissingField("space"))
	}
	if cf.Org == "" {
		errs = append(errs, linterrors.NewMissingField("org"))
	}
	if cf.TestDomain == "" {
		_, found := defaults.DefaultValues.CfTestDomains[cf.API]
		if cf.API != "" && !found {
			errs = append(errs, linterrors.NewMissingField("testDomain"))
		}
	}

	if cf.Retries < 0 || cf.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if strings.HasPrefix(cf.Manifest, "../artifacts/") {
		warnings = append(warnings, linterrors.NewFileError(cf.Manifest, "this file must be saved as an artifact in a previous task"))
		if len(cf.PrePromote) > 0 {
			errs = append(errs, linterrors.NewInvalidField("pre_promote", "if you are using generated manifest you cannot have pre promote tasks"))

		}
	} else {

		if err := filechecker.CheckFile(fs, cf.Manifest, false); err != nil {
			errs = append(errs, err)
		}
	}

	for _, preStartCommand := range cf.PreStart {
		if !strings.HasPrefix(preStartCommand, "cf ") {
			errs = append(errs, linterrors.NewInvalidField("pre_start", fmt.Sprintf("only cf commands are allowed: %s", preStartCommand)))
		}
	}

	return errs, warnings
}
