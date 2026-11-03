package linters

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintObservabilityProvisioningTask(task manifest.ObservabilityProvisioning, fs afero.Afero) (errs []error) {
	exists, err := fs.DirExists(task.SLOs)
	if err != nil {
		return []error{err}
	}

	if !exists {
		return []error{ErrSLOsDirectoryNotFound(task.SLOs)}
	}

	return nil
}
