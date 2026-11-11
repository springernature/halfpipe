package defaults

import "github.com/springernature/halfpipe/manifest"

func deployKateeDefaulter(original manifest.DeployKatee, defaults Defaults, man manifest.Manifest) (updated manifest.DeployKatee) {
	updated = original

	if updated.VelaManifest == "" {
		updated.VelaManifest = defaults.Katee.VelaManifest
	}

	if ns := original.Vars["KATEE_NAMESPACE"]; ns != "" {
		// The old Concourse task let vars override the namespace, and pipelines rely on it.
		updated.Namespace = ns
	} else if original.Namespace == "" {
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

	if original.Retries == 0 {
		updated.Retries = 1
	}

	return updated
}
