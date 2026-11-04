package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintObservabilityProvisioningTask(task manifest.ObservabilityProvisioning, fs afero.Afero) (errs []error) {
	resources := []string{task.SLOs, task.MetricRules}
	anyDefined := false

	for _, res := range resources {
		if res == "" {
			continue
		}

		anyDefined = true

		if err := checkFolderExists(res, fs); err != nil {
			errs = append(errs, err)
		}
	}

	if !anyDefined {
		return []error{ErrNoObservabilityResourcesDefined}
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
