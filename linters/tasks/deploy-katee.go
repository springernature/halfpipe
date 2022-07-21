package tasks

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

func LintDeployKateeTask(task manifest.DeployKatee, man manifest.Manifest, fs afero.Afero) (errs []error, warnings []error) {
	return errs, warnings
}
