package defaults

import "github.com/springernature/halfpipe/config"

var Concourse = Defaults{
	RepoPrivateKey: config.VaultSecrets.GitHubPrivateKey,
	ShallowClone:   false,
	CF: CFDefaults{
		SnPaaS: CFSnPaaS{
			Username: config.VaultSecrets.CFSnPaaSUsername,
			Password: config.VaultSecrets.CFSnPaaSPassword,
			Org:      config.VaultSecrets.CFSnPaaSOrg,
			API:      config.VaultSecrets.CFSnPaaSAPI,
		},
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.snpaas.eu":         "springernature.app",
			config.VaultSecrets.CFSnPaaSAPI: "springernature.app",
		},
		Version: "cf7",
	},
	Katee: KateeDefaults{
		VelaManifest:  "vela.yaml",
		Tag:           "version",
		CheckInterval: 5,
		MaxChecks:     60,
	},
	Docker: DockerDefaults{
		Username:       "oauth2accesstoken",
		Password:       config.VaultSecrets.GARToken,
		ComposeService: "app",
		ComposeFile:    []string{"docker-compose.yml"},
		FilePath:       "Dockerfile",
	},
	Artifactory: ArtifactoryDefaults{
		Username: config.VaultSecrets.ArtifactoryUsername,
		Password: config.VaultSecrets.ArtifactoryPassword,
		URL:      config.VaultSecrets.ArtifactoryURL,
	},
	Concourse: ConcourseDefaults{
		URL:      config.VaultSecrets.ConcourseURL,
		Username: config.VaultSecrets.ConcourseUsername,
		Password: config.VaultSecrets.ConcoursePassword,
	},
	MarkLogic: MarkLogicDefaults{
		Username: config.VaultSecrets.MarkLogicUsername,
		Password: config.VaultSecrets.MarkLogicPassword,
	},
	Timeout: "1h",
	Buildpack: BuildpackDefaults{
		Builder: "paketobuildpacks/builder-jammy-buildpackless-base",
	},
	AWSDocker: AWSDockerDefaults{
		Region:          "cn-northwest-1",
		AccessKeyID:     config.VaultSecrets.AWSECRAccessKeyID,
		SecretAccessKey: config.VaultSecrets.AWSECRSecretAccessKey,
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
			Username: config.VaultSecrets.CFSnPaaSUsername,
			Password: config.VaultSecrets.CFSnPaaSPassword,
			Org:      config.VaultSecrets.CFSnPaaSOrg,
			API:      config.VaultSecrets.CFSnPaaSAPI,
		},
		ManifestPath: "manifest.yml",
		TestDomains: map[string]string{
			"https://api.snpaas.eu":         "springernature.app",
			config.VaultSecrets.CFSnPaaSAPI: "springernature.app",
		},
		Version: "cf7",
	},
	Katee: KateeDefaults{
		VelaManifest:  "vela.yaml",
		Tag:           "version",
		CheckInterval: 5,
		MaxChecks:     60,
	},
	Buildpack: BuildpackDefaults{
		Builder: "paketobuildpacks/builder-jammy-buildpackless-base",
	},
	AWSDocker: AWSDockerDefaults{
		Region:          "cn-northwest-1",
		AccessKeyID:     config.VaultSecrets.AWSECRAccessKeyID,
		SecretAccessKey: config.VaultSecrets.AWSECRSecretAccessKey,
	},
}
