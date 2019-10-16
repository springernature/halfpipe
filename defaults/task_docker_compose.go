package defaults

import "github.com/springernature/halfpipe/manifest"

func dockerComposeDefaulter(original manifest.DockerCompose, defaults Defaults) (updated manifest.DockerCompose) {
	updated = original

	if updated.Service == "" {
		updated.Service = defaults.DockerComposeService
	}

	return
}
