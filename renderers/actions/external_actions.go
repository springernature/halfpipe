package actions

import (
	"reflect"
	"strings"
)

// ExternalAction represents a GitHub Action with a pinned SHA and version tag.
type ExternalAction struct {
	Ref     string // Full action reference (e.g., "actions/checkout@sha")
	Version string // Version tag (e.g., "v6.0.2") - only for SHA-pinned actions
}

var ExternalActions = struct {
	Buildpack            ExternalAction
	Checkout             ExternalAction
	CreateGitHubAppToken ExternalAction
	DeployCF             ExternalAction
	DeployKatee          ExternalAction
	DockerLogin          ExternalAction
	DockerPush           ExternalAction
	DownloadArtifact     ExternalAction
	RepositoryDispatch   ExternalAction
	Slack                ExternalAction
	Teams                ExternalAction
	UploadArtifact       ExternalAction
	Vault                ExternalAction
}{
	Buildpack:            ExternalAction{Ref: "springernature/ee-action-buildpack@v1"},
	Checkout:             ExternalAction{Ref: "actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0", Version: "v7.0.0"},
	CreateGitHubAppToken: ExternalAction{Ref: "actions/create-github-app-token@bcd2ba49218906704ab6c1aa796996da409d3eb1", Version: "v3.2.0"},
	DeployCF:             ExternalAction{Ref: "springernature/ee-action-deploy-cf@v1"},
	DeployKatee:          ExternalAction{Ref: "springernature/ee-action-deploy-katee@v1"},
	DockerLogin:          ExternalAction{Ref: "docker/login-action@650006c6eb7dba73a995cc03b0b2d7f5ca915bee", Version: "v4.2.0"},
	DockerPush:           ExternalAction{Ref: "springernature/ee-action-docker-push@v1"},
	DownloadArtifact:     ExternalAction{Ref: "actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c", Version: "v8.0.1"},
	RepositoryDispatch:   ExternalAction{Ref: "peter-evans/repository-dispatch@28959ce8df70de7be546dd1250a005dd32156697", Version: "v4.0.1"},
	Slack:                ExternalAction{Ref: "slackapi/slack-github-action@45a88b9581bfab2566dc881e2cd66d334e621e2c", Version: "v3.0.3"},
	Teams:                ExternalAction{Ref: "springernature/ee-action-ms-teams@v1"},
	UploadArtifact:       ExternalAction{Ref: "actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a", Version: "v7.0.1"},
	Vault:                ExternalAction{Ref: "hashicorp/vault-action@892a26828f195e65540a40b4768ae4571f51ebfc", Version: "v4.0.0"},
}

// GetAllSHAPinnedActions returns a map of SHA to version for all SHA-pinned actions.
// This is used for adding version comments to rendered YAML.
func GetAllSHAPinnedActions() map[string]string {
	result := make(map[string]string)
	v := reflect.ValueOf(ExternalActions)
	for i := 0; i < v.NumField(); i++ {
		a := v.Field(i).Interface().(ExternalAction)
		if a.Version != "" {
			if _, sha, found := strings.Cut(a.Ref, "@"); found {
				result[sha] = a.Version
			}
		}
	}
	return result
}
