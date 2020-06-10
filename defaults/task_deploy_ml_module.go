package defaults

import "github.com/springernature/halfpipe/manifest"

func deployMlModuleDefaulter(original manifest.DeployMLModules, defaults Defaults) (updated manifest.DeployMLModules) {
	updated = original

	if updated.Username == "" {
		updated.Username = defaults.MarkLogicUsername
	}

	if updated.Password == "" {
		updated.Password = defaults.MarkLogicPassword
	}

	return updated
}
