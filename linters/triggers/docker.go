package triggers

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerTrigger(docker manifest.DockerTrigger) (errs []error, warnings []error) {
	if docker.Image == "" {
		errs = append(errs, linterrors.NewMissingField("image"))
	}

	return errs, warnings
}
