package defaults

import "github.com/springernature/halfpipe/config"

var commonCF = CFDefaults{
	ManifestPath: "manifest.yml",
	Version:      "cf7",
	SnPaaS: CFSnPaaS{
		Username: config.VaultSecrets.CFSnPaaSUsername,
		Password: config.VaultSecrets.CFSnPaaSPassword,
		Org:      config.VaultSecrets.CFSnPaaSOrg,
		API:      config.VaultSecrets.CFSnPaaSAPI,
	},
	TestDomains: map[string]string{
		"https://api.snpaas.eu":         "springernature.app",
		config.VaultSecrets.CFSnPaaSAPI: "springernature.app",
	},
}

var commonKatee = KateeDefaults{
	VelaManifest:  "vela.yaml",
	Tag:           "version",
	CheckInterval: 5,
	MaxChecks:     60,
}

var commonDocker = DockerDefaults{
	FilePath:       "Dockerfile",
	ComposeFile:    []string{"docker-compose.yml"},
	ComposeService: "app",
}

var commonBuildpack = BuildpackDefaults{
	Builder: "paketobuildpacks/builder-jammy-buildpackless-base",
}

var commonAWSDocker = AWSDockerDefaults{
	Region:          "cn-northwest-1",
	AccessKeyID:     config.VaultSecrets.AWSECRAccessKeyID,
	SecretAccessKey: config.VaultSecrets.AWSECRSecretAccessKey,
}

var commonMarkLogic = MarkLogicDefaults{
	Username: config.VaultSecrets.MarkLogicUsername,
	Password: config.VaultSecrets.MarkLogicPassword,
}

var Concourse = Defaults{
	RepoPrivateKey: config.VaultSecrets.GitHubPrivateKey,
	ShallowClone:   false,
	Timeout:        "1h",
	CF:             commonCF,
	Katee:          commonKatee,
	Docker: DockerDefaults{
		Username:       "oauth2accesstoken",
		Password:       config.VaultSecrets.GARToken,
		FilePath:       commonDocker.FilePath,
		ComposeFile:    commonDocker.ComposeFile,
		ComposeService: commonDocker.ComposeService,
	},
	Buildpack: commonBuildpack,
	AWSDocker: commonAWSDocker,
	MarkLogic: commonMarkLogic,
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
}

var Actions = Defaults{
	ShallowClone: true,
	CF:           commonCF,
	Katee:        commonKatee,
	Docker:       commonDocker,
	Buildpack:    commonBuildpack,
	AWSDocker:    commonAWSDocker,
	MarkLogic:    commonMarkLogic,
}
