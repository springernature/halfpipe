package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployKateeTask(task manifest.DeployKatee, fs afero.Afero) (errs []error, warnings []error) {
	if task.ApplicationName == "" {
		errs = append(errs, linterrors.NewMissingField("applicationName"))
	}

	if task.Retries < 0 || task.Retries > 5 {
		errs = append(errs, linterrors.NewInvalidField("retries", "must be between 0 and 5"))
	}

	if err := filechecker.CheckFile(fs, task.VelaAppFile, false); err != nil {
		errs = append(errs, err)
	}

	return errs, warnings
}
