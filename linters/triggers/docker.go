package triggers

import (
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func LintDockerTrigger(docker manifest.DockerTrigger, deprecatedDockerRegistries []string) (errs []error, warnings []error) {
	if docker.Image == "" {
		errs = append(errs, linterrors.NewMissingField("image"))
	}

	for _, hostname := range deprecatedDockerRegistries {
		if strings.HasPrefix(docker.Image, hostname) {
			warnings = append(warnings, linterrors.NewDeprecatedDockerRegistryError(hostname))
		}
	}

	return errs, warnings
}
