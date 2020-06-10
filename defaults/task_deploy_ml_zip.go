package defaults

import "github.com/springernature/halfpipe/manifest"

func deployMlZipDefaulter(original manifest.DeployMLZip, defaults Defaults) (updated manifest.DeployMLZip) {
	updated = original

	if updated.Username == "" {
		updated.Username = defaults.MarkLogicUsername
	}

	if updated.Password == "" {
		updated.Password = defaults.MarkLogicPassword
	}

	return updated
}
