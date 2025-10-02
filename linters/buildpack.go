package linters

import (
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func LintBuildpackTask(task manifest.Buildpack) (errs []error) {

	if task.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	if task.Buildpacks == "" {
		errs = append(errs, NewErrMissingField("buildpacks"))
	}

	if !strings.HasPrefix(task.Builder, "paketobuildpacks/") {
		errs = append(errs, ErrInvalidBuilder)

	}

	return errs
}
