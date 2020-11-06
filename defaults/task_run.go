package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func runDefaulter(original manifest.Run, defaults Defaults) (updated manifest.Run) {
	updated = original
	if strings.HasPrefix(updated.Docker.Image, config.DockerRegistry) {
		updated.Docker.Username = defaults.Docker.Username
		updated.Docker.Password = defaults.Docker.Password
	}

	return updated
}
