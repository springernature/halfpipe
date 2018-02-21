package defaults

// this file could be overwritten before build in Concourse

var DefaultValues = Defaults{
	RepoPrivateKey: "((github.deploy-key))",
	CfUsername:     "((cloudfoundry.username))",
	CfPassword:     "((cloudfoundry.password))",
	CfManifest:     "manifest.yml",
}
