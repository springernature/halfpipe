package linters

import (
	"strings"

	"github.com/springernature/halfpipe/manifest"
)

func LintBuildpackTask(task manifest.Buildpack) (errs []error) {

	if task.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	if len(task.Buildpacks) == 0 {
		errs = append(errs, NewErrMissingField("buildpacks"))
	}

	if !strings.HasPrefix(task.Builder, "paketobuildpacks/") {
		errs = append(errs, ErrInvalidBuilder)

	}

	return errs
}
