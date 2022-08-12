package defaults

import (
	"strings"

	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func dockerPushDefaulter(original manifest.DockerPush, man manifest.Manifest, defaults Defaults) (updated manifest.DockerPush) {
	updated = original

	if man.Platform.IsConcourse() && strings.HasPrefix(updated.Image, config.DockerRegistry) {
		updated.Username = defaults.Docker.Username
		updated.Password = defaults.Docker.Password
	}

	if updated.DockerfilePath == "" {
		updated.DockerfilePath = defaults.Docker.FilePath
	}

	if original.ImageScanSeverity == "" {
		updated.ImageScanSeverity = defaults.Docker.ImageScanSeverity
	}

	return updated
}
