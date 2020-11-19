package defaults

import "github.com/springernature/halfpipe/manifest"

func deployMlZipDefaulter(original manifest.DeployMLZip, defaults Defaults) (updated manifest.DeployMLZip) {
	updated = original

	if updated.Username == "" {
		updated.Username = defaults.MarkLogic.Username
	}

	if updated.Password == "" {
		updated.Password = defaults.MarkLogic.Password
	}

	return updated
}
