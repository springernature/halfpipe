package defaults

import "github.com/springernature/halfpipe/manifest"

func deployCfDefaulter(original manifest.DeployCF, defaults Defaults, man manifest.Manifest) (updated manifest.DeployCF) {
	updated = original

	if updated.API == defaults.CfAPISnPaas {
		if updated.Org == "" {
			updated.Org = defaults.CfOrgSnPaas
		}
		if updated.Username == "" {
			updated.Username = defaults.CfUsernameSnPaas
		}
		if updated.Password == "" {
			updated.Password = defaults.CfPasswordSnPaas
		}
	} else {
		if updated.Org == "" {
			updated.Org = man.Team
		}
		if updated.Username == "" {
			updated.Username = defaults.CfUsername
		}
		if updated.Password == "" {
			updated.Password = defaults.CfPassword
		}
	}

	if updated.Manifest == "" {
		updated.Manifest = defaults.CfManifest
	}

	if updated.TestDomain == "" {
		if domain, ok := defaults.CfTestDomains[updated.API]; ok {
			updated.TestDomain = domain
		}
	}

	return updated
}
