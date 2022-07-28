package defaults

import "github.com/springernature/halfpipe/manifest"

func deployKateeDefaulter(original manifest.DeployKatee, defaults Defaults, man manifest.Manifest) (updated manifest.DeployKatee) {
	updated = original

	if updated.VelaAppFile == "" {
		updated.VelaAppFile = defaults.Katee.VelaAppFile
	}

	if updated.Tag == "" {
		updated.Tag = defaults.Katee.Tag
	}

	return updated
}
