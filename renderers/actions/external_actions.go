package actions

var ExternalActions = struct {
	Buildpack          string
	Checkout           string
	DeployCF           string
	DeployKatee        string
	DockerLogin        string
	DockerPush         string
	DownloadArtifact   string
	RepositoryDispatch string
	Slack              string
	Teams              string
	UploadArtifact     string
	Vault              string
}{
	Buildpack:          "springernature/ee-action-buildpack@v1",
	Checkout:           "actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd", // v6.0.2
	DeployCF:           "springernature/ee-action-deploy-cf@v1",
	DeployKatee:        "springernature/ee-action-deploy-katee@v1",
	DockerLogin:        "docker/login-action@b45d80f862d83dbcd57f89517bcf500b2ab88fb2", // v4.0.0
	DockerPush:         "springernature/ee-action-docker-push@v1",
	DownloadArtifact:   "actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c",       // v8.0.1
	RepositoryDispatch: "peter-evans/repository-dispatch@28959ce8df70de7be546dd1250a005dd32156697", // v4.0.1
	Slack:              "slackapi/slack-github-action@af78098f536edbc4de71162a307590698245be95",    // v3.0.1
	Teams:              "jdcargile/ms-teams-notification@28e5ca976c053d54e2b852f3f38da312f35a24fc", // v1.4
	UploadArtifact:     "actions/upload-artifact@bbbca2ddaa5d8feaa63e36b76fdaad77386f024f",         // v7.0.0
	Vault:              "hashicorp/vault-action@4c06c5ccf5c0761b6029f56cfb1dcf5565918a3b",          // v3.4.0
}
