package actions

var ExternalActions = struct {
	AWSCredentials     string
	AWSCLI             string
	AWSECRLogin        string
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
	AWSCredentials:     "aws-actions/configure-aws-credentials@ececac1a45f3b08a01d2dd070d28d111c5fe175a", // v4
	AWSCLI:             "unfor19/install-aws-cli-action@e8b481e524a99f37fbd39fdc1dcb3341ab091571",        // v1
	AWSECRLogin:        "aws-actions/amazon-ecr-login@v2",
	Buildpack:          "springernature/ee-action-buildpack@v1",
	Checkout:           "actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8", // v6
	DeployCF:           "springernature/ee-action-deploy-cf@v1",
	DeployKatee:        "springernature/ee-action-deploy-katee@v1",
	DockerLogin:        "docker/login-action@9780b0c442fbb1117ed29e0efdff1e18412f7567", // v3
	DockerPush:         "springernature/ee-action-docker-push@v1",
	DownloadArtifact:   "actions/download-artifact@37930b1c2abaa49bbe596cd826c3c89aef350131",       // v7
	RepositoryDispatch: "peter-evans/repository-dispatch@28959ce8df70de7be546dd1250a005dd32156697", // v4
	Slack:              "slackapi/slack-github-action@91efab103c0de0a537f72a35f6b8cda0ee76bf0a",    //v2
	Teams:              "jdcargile/ms-teams-notification@28e5ca976c053d54e2b852f3f38da312f35a24fc", // v1
	UploadArtifact:     "actions/upload-artifact@b7c566a772e6b6bfb58ed0dc250532a479d7789f",         // v6
	Vault:              "hashicorp/vault-action@4c06c5ccf5c0761b6029f56cfb1dcf5565918a3b",          // v3
}
