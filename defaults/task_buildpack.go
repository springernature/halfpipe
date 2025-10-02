package defaults

import "github.com/springernature/halfpipe/manifest"

func buildpackDefaulter(original manifest.Buildpack, defaults Defaults) (updated manifest.Buildpack) {
	updated = original

	if original.Builder == "" {
		updated.Builder = defaults.Buildpack.Builder
	}

	return updated
}
