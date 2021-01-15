package defaults

var Concourse = Defaults{
	RepoPrivateKey: "((halfpipe-github.private_key))",
	CF: CFDefaults{
		SnPaaS: CFSnPaaS{
			Username: "((cloudfoundry.username-snpaas))",
			Password: "((cloudfoundry.password-snpaas))",
			Org:      "((cloudfoundry.org-snpaas))",
			API:      "((cloudfoundry.api-snpaas))",
		},
		OnPrem: CFOnPrem{
			Username: "((cloudfoundry.username))",
			Password: "((cloudfoundry.password))",
		},
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.dev.cf.springer-sbm.com": "dev.cf.private.springer.com",
			"((cloudfoundry.api-dev))":            "dev.cf.private.springer.com",

			"https://api.live.cf.springer-sbm.com": "live.cf.private.springer.com",
			"((cloudfoundry.api-live))":            "live.cf.private.springer.com",

			"https://api.snpaas.eu":       "springernature.app",
			"((cloudfoundry.api-snpaas))": "springernature.app",
		},
		Version: "cf6",
	},
	Docker: DockerDefaults{
		Username:       "_json_key",
		Password:       "((halfpipe-gcr.private_key))",
		ComposeService: "app",
		ComposeFile:    "docker-compose.yml",
		FilePath:       "Dockerfile",
	},
	Artifactory: ArtifactoryDefaults{
		Username: "((artifactory.username))",
		Password: "((artifactory.password))",
		URL:      "((artifactory.url))",
	},
	Concourse: ConcourseDefaults{
		URL:      "((concourse.url))",
		Username: "((concourse.username))",
		Password: "((concourse.password))",
	},
	MarkLogic: MarkLogicDefaults{
		Username: "((halfpipe-ml-deploy.username))",
		Password: "((halfpipe-ml-deploy.password))",
	},
	Timeout: "1h",
}

var Actions = Defaults{
	RepoPrivateKey: "this cannot be empty due to linter",

	Docker: DockerDefaults{
		Username:       "_json_key",
		Password:       "${{ secrets.EE_GCR_PRIVATE_KEY }}",
		ComposeService: "app",
		ComposeFile:    "docker-compose.yml",
		FilePath:       "Dockerfile",
	},

	CF: CFDefaults{
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.dev.cf.springer-sbm.com": "dev.cf.private.springer.com",
			"((cloudfoundry.api-dev))":            "dev.cf.private.springer.com",

			"https://api.live.cf.springer-sbm.com": "live.cf.private.springer.com",
			"((cloudfoundry.api-live))":            "live.cf.private.springer.com",

			"https://api.snpaas.eu":       "springernature.app",
			"((cloudfoundry.api-snpaas))": "springernature.app",
		},
		Version: "cf7",
	},

	RemoveUpdatePipelineFeature: true,
}
