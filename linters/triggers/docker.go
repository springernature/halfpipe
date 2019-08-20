package triggers

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintDockerTrigger(docker manifest.DockerTrigger) (errs []error, warnings []error) {
	if docker.Image == "" {
		errs = append(errs, errors.NewMissingField("image"))
	}

	return
}
