package defaults

import (
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func dockerPushDefaulter(original manifest.DockerPush, man manifest.Manifest, defaults Defaults) (updated manifest.DockerPush) {
	updated = original

	if strings.HasPrefix(updated.Image, config.DockerRegistry) {
		updated.Username = defaults.DockerUsername
		updated.Password = defaults.DockerPassword
	}

	if updated.DockerfilePath == "" {
		updated.DockerfilePath = defaults.DockerfilePath
	}

	if original.Tag == "" {
		if man.FeatureToggles.Versioned() {
			updated.Tag = "version"
		} else {
			updated.Tag = "gitref"
		}
	}

	return updated
}
