package defaults

import "github.com/springernature/halfpipe/manifest"

func dockerComposeDefaulter(original manifest.DockerCompose, defaults Defaults) (updated manifest.DockerCompose) {
	updated = original

	if len(original.ComposeFiles) == 0 {
		updated.ComposeFiles = defaults.Docker.ComposeFile
	}

	if original.Service == "" {
		updated.Service = defaults.Docker.ComposeService
	}

	return updated
}
