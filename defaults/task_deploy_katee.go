package defaults

import "github.com/springernature/halfpipe/manifest"

func deploygit checkout -b <branch-name>er(original manifest.DeployKatee, defaults Defaults, man manifest.Manifest) (updated manifest.DeployKatee) {
	updated = original

	if updated.VelaManifest == "" {
		updated.VelaManifest = defaults.Katee.VelaManifest
	}

	if original.Tag == "" {
		if man.Platform.IsActions() || man.FeatureToggles.UpdatePipeline() {
			updated.Tag = "version"
		} else {
			updated.Tag = "gitref"
		}
	}

	if original.Namespace == "" {
		updated.Namespace = "katee-" + man.Team
	}

	if original.Environment == "" {
		updated.Environment = man.Team
	}

	if original.PlatformVersion == "" {
		updated.PlatformVersion = "v1"
	}

	return updated
}
