package defaults

var DefaultValues = Defaults{
	RepoPrivateKey: "((github.private_key))",
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
	},
	DockerUsername:       "_json_key",
	DockerPassword:       "((gcr.private_key))",
	DockerComposeService: "app",
}
