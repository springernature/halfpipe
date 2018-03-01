package defaults

var DefaultValues = Defaults{
	RepoPrivateKey: "((github.private_key))",
	CfUsername:     "((cloudfoundry.username))",
	CfPassword:     "((cloudfoundry.password))",
	CfManifest:     "manifest.yml",
	DockerUsername: "_json_key",
	DockerPassword: "((gcr.private_key))",
}
