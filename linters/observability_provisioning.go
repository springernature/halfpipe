package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintObservabilityProvisioningTask(task manifest.ObservabilityProvisioning, fs afero.Afero) (errs []error) {
	if err := checkFolderExists(task.SLOs, fs); err != nil && task.SLOs != "" {
		errs = append(errs, err)
	}

	if err := checkFolderExists(task.MetricRules, fs); err != nil && task.MetricRules != "" {
		errs = append(errs, err)
	}

	return errs
}

func checkFolderExists(folder string, fs afero.Afero) error {
	exists, err := fs.DirExists(folder)
	if err != nil {
		return err
	}
	if !exists {
		return NewErrFolderNotFound(folder)
	}
	return nil
}
