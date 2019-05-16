package defaults

var DefaultValues = Defaults{
	RepoPrivateKey: "((halfpipe-github.private_key))",
	CfUsername:     "((cloudfoundry.username))",
	CfPassword:     "((cloudfoundry.password))",
	CfManifest:     "manifest.yml",
	CfTestDomains: map[string]string{
		"https://api.dev.cf.springer-sbm.com": "dev.cf.private.springer.com",
		"((cloudfoundry.api-dev))":            "dev.cf.private.springer.com",

		"https://api.live.cf.springer-sbm.com": "live.cf.private.springer.com",
		"((cloudfoundry.api-live))":            "live.cf.private.springer.com",

		"https://api.europe-west1.cf.gcp.springernature.io": "apps.gcp.springernature.io",
		"((cloudfoundry.api-gcp))":                          "apps.gcp.springernature.io",

		"https://api.snpaas.eu":       "springernature.app",
		"((cloudfoundry.api-snpaas))": "springernature.app",
	},
	CfUsernameSnPaas:     "((cloudfoundry.username-snpaas))",
	CfPasswordSnPaas:     "((cloudfoundry.password-snpaas))",
	CfOrgSnPaas:          "((cloudfoundry.org-snpaas))",
	CfAPISnPaas:          "((cloudfoundry.api-snpaas))",
	DockerUsername:       "_json_key",
	DockerPassword:       " ((halfpipe-gcr.private_key))",
	DockerComposeService: "app",

	ArtifactoryUsername: "((artifactory.username))",
	ArtifactoryPassword: "((artifactory.password))",
	ArtifactoryURL:      "((artifactory.url))",

	Timeout: "1h",
}
