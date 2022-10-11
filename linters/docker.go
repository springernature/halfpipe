package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerTrigger(docker manifest.DockerTrigger) (errs []error) {
	if docker.Image == "" {
		errs = append(errs, NewErrMissingField("image"))
	}

	return errs
}
