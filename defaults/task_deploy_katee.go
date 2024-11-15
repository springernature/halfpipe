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

	if original.Namespace == "" {
		updated.Namespace = "katee-" + man.Team
	}

	if original.DeploymentCheckTimeout > 0 {
		updated.MaxChecks = original.DeploymentCheckTimeout
	}

	if updated.MaxChecks == 0 {
		updated.MaxChecks = defaults.Katee.MaxChecks
	}

	if original.CheckInterval == 0 {
		updated.CheckInterval = defaults.Katee.CheckInterval
	}

	return updated
}
