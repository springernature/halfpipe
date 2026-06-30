package defaults

import "github.com/springernature/halfpipe/manifest"

func deployCfDefaulter(original manifest.DeployCF, defaults Defaults, man manifest.Manifest) (updated manifest.DeployCF) {
	updated = original
	if updated.API == "" || updated.API == "((cloudfoundry.api-snpaas))" {
		updated.API = defaults.CF.SnPaaS.API
	}

	if updated.API == defaults.CF.SnPaaS.API {
		if updated.Org == "" || updated.Org == "((cloudfoundry.org-snpaas))" {
			updated.Org = defaults.CF.SnPaaS.Org
		}
		if updated.Username == "" || updated.Username == "((cloudfoundry.username-snpaas))" {
			updated.Username = defaults.CF.SnPaaS.Username
		}
		if updated.Password == "" || updated.Password == "((cloudfoundry.password-snpaas))" {
			updated.Password = defaults.CF.SnPaaS.Password
		}
	} else {
		if updated.Org == "" {
			updated.Org = man.Team
		}
	}

	if updated.Manifest == "" {
		updated.Manifest = defaults.CF.ManifestPath
	}

	if updated.TestDomain == "" {
		if domain, ok := defaults.CF.TestDomains[updated.API]; ok {
			updated.TestDomain = domain
		}
	}

	if updated.CliVersion == "" {
		updated.CliVersion = defaults.CF.Version
	}

	if original.Retries == 0 {
		updated.Retries = 1
	}

	return updated
}
