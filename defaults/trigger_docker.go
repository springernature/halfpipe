package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func defaultDockerTrigger(original manifest.DockerTrigger, defaults Defaults) (updated manifest.DockerTrigger) {
	updated = original
	if strings.HasPrefix(updated.Image, config.DockerRegistry) {
		updated.Username = defaults.Docker.Username
		updated.Password = defaults.Docker.Password
	}
	return updated
}
