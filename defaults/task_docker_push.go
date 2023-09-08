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

	if updated.ScanTimeout == 0 {
		updated.ScanTimeout = 15
	}

	if len(original.Platforms) == 0 {
		updated.Platforms = []string{"linux/amd64"}
	}

	if updated.Secrets == nil {
		updated.Secrets = make(manifest.Vars)
	}

	if man.Platform.IsConcourse() {
		updated.Secrets["ARTIFACTORY_URL"] = defaults.Artifactory.URL
		updated.Secrets["ARTIFACTORY_USERNAME"] = defaults.Artifactory.Username
		updated.Secrets["ARTIFACTORY_PASSWORD"] = defaults.Artifactory.Password
	}
	if man.Platform.IsActions() {
		updated.Secrets["ARTIFACTORY_URL"] = "${{ secrets.EE_ARTIFACTORY_URL }}"
		updated.Secrets["ARTIFACTORY_USERNAME"] = "${{ secrets.EE_ARTIFACTORY_USERNAME }}"
		updated.Secrets["ARTIFACTORY_PASSWORD"] = "${{ secrets.EE_ARTIFACTORY_PASSWORD }}"
	}

	return updated
}
