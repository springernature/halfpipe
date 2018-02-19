package defaults

// this file could be overwritten before build in Concourse

var DefaultValues = Defaults{
	RepoPrivateKey: "((deploy-key))",
	CfUsername:     "((cf-credentials.username))",
	CfPassword:     "((cf-credentials.password))",
	CfManifest:     "manifest.yml",
}
