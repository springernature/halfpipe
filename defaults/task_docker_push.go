package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func dockerPushDefaulter(original manifest.DockerPush, defaults Defaults) (updated manifest.DockerPush) {
	updated = original

	if strings.HasPrefix(updated.Image, config.DockerRegistry) {
		updated.Username = defaults.DockerUsername
		updated.Password = defaults.DockerPassword
	}

	if updated.DockerfilePath == "" {
		updated.DockerfilePath = "Dockerfile"
	}

	return updated
}
