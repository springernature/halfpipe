package defaults

import "github.com/springernature/halfpipe/manifest"

func deployKateeDefaulter(original manifest.DeployKatee, defaults Defaults, man manifest.Manifest) (updated manifest.DeployKatee) {
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

	if original.Team == "" {
		updated.Team = man.Team
	}

	return updated
}
