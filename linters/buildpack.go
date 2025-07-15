package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintBuildpackTask(task manifest.Buildpack) (errs []error) {

	if task.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	if task.Buildpacks == "" {
		errs = append(errs, NewErrMissingField("buildpacks"))
	}

	return errs
}
