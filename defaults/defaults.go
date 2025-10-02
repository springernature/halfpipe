package defaults

var Concourse = Defaults{
	RepoPrivateKey: "((halfpipe-github.private_key))",
	ShallowClone:   false,
	CF: CFDefaults{
		SnPaaS: CFSnPaaS{
			Username: "((cloudfoundry.username-snpaas))",
			Password: "((cloudfoundry.password-snpaas))",
			Org:      "((cloudfoundry.org-snpaas))",
			API:      "((cloudfoundry.api-snpaas))",
		},
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.snpaas.eu":       "springernature.app",
			"((cloudfoundry.api-snpaas))": "springernature.app",
		},
		Version: "cf7",
	},
	Katee: KateeDefaults{
		VelaManifest:  "vela.yaml",
		Tag:           "version",
		CheckInterval: 2,
		MaxChecks:     60,
	},
	Docker: DockerDefaults{
		Username:       "_json_key",
		Password:       "((halfpipe-gcr.private_key))",
		ComposeService: "app",
		ComposeFile:    []string{"docker-compose.yml"},
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
	Buildpack: BuildpackDefaults{
		Builder: "paketobuildpacks/builder-jammy-buildpackless-full",
	},
}

var Actions = Defaults{
	ShallowClone: true,
	Docker: DockerDefaults{
		ComposeService: "app",
		ComposeFile:    []string{"docker-compose.yml"},
		FilePath:       "Dockerfile",
	},

	CF: CFDefaults{
		SnPaaS: CFSnPaaS{
			Username: "((cloudfoundry.username-snpaas))",
			Password: "((cloudfoundry.password-snpaas))",
			Org:      "((cloudfoundry.org-snpaas))",
			API:      "((cloudfoundry.api-snpaas))",
		},
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.snpaas.eu":       "springernature.app",
			"((cloudfoundry.api-snpaas))": "springernature.app",
		},
		Version: "cf7",
	},
	Katee: KateeDefaults{
		VelaManifest:  "vela.yaml",
		Tag:           "version",
		CheckInterval: 2,
		MaxChecks:     60,
	},
	Buildpack: BuildpackDefaults{
		Builder: "paketobuildpacks/builder-jammy-buildpackless-full",
	},
}
