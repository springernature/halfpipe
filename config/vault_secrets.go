package config

import (
	"fmt"
)

var VaultSecrets = struct {
	GitHubPrivateKey, GithubStatusesToken, HalfpipeBotAppID, HalfpipeBotInstallationID, HalfpipeBotPrivateKey,
	GARToken, GCPArtifactsToken, GCRPrivateKey, GCRPrivateKeyBase64, VersionBucket, VersionJSONKey, ArtifactsBucket,
	VaultAddr, VaultRoleID, VaultSecretID,
	ConcourseURL, ConcoursePassword, ConcourseTeam, ConcourseUsername,
	CFSnPaaSUsername, CFSnPaaSPassword, CFSnPaaSOrg, CFSnPaaSAPI,
	ArtifactoryUsername, ArtifactoryPassword, ArtifactoryURL,
	MarkLogicUsername, MarkLogicPassword,
	AWSECRAccessKeyID, AWSECRSecretAccessKey,
	SlackToken string
	KateeKey func(string, string, string) string
}{
	// github
	GitHubPrivateKey:          "((halfpipe-github.private_key))",
	GithubStatusesToken:       "((halfpipe-github.statuses-token))",
	HalfpipeBotAppID:          "((halfpipe-bot.app_id))",
	HalfpipeBotInstallationID: "((halfpipe-bot.installation_id))",
	HalfpipeBotPrivateKey:     "((halfpipe-bot.private_key))",

	// gcp
	GARToken:            "((gcp:platform-gar/token.token))",
	GCPArtifactsToken:   "((gcp:platform-artifacts/token.token))",
	GCRPrivateKey:       "((halfpipe-gcr.private_key))",
	GCRPrivateKeyBase64: "((halfpipe-gcr.private_key_base64))",
	VersionBucket:       "((halfpipe-semver.bucket))",
	VersionJSONKey:      "((halfpipe-semver.private_key))",
	ArtifactsBucket:     "((halfpipe-artifacts.bucket))",

	// vault
	VaultAddr:     "((platform/team-ro-app-role.vault_addr))",
	VaultRoleID:   "((platform/team-ro-app-role.vault_approle_id))",
	VaultSecretID: "((platform/team-ro-app-role.vault_approle_secret_id))",

	// concourse
	ConcourseURL:      "((platform/concourse.url))",
	ConcoursePassword: "((platform/concourse.password))",
	ConcourseTeam:     "((platform/concourse.team))",
	ConcourseUsername: "((platform/concourse.username))",

	// cloudfoundry snpaas
	CFSnPaaSUsername: "((platform/cloudfoundry.username-snpaas))",
	CFSnPaaSPassword: "((platform/cloudfoundry.password-snpaas))",
	CFSnPaaSOrg:      "((platform/cloudfoundry.org-snpaas))",
	CFSnPaaSAPI:      "((platform/cloudfoundry.api-snpaas))",

	// artifactory
	ArtifactoryUsername: "((artifactory.username))",
	ArtifactoryPassword: "((artifactory.password))",
	ArtifactoryURL:      "((artifactory.url))",

	// marklogic
	MarkLogicUsername: "((halfpipe-ml-deploy.username))",
	MarkLogicPassword: "((halfpipe-ml-deploy.password))",

	// aws
	AWSECRAccessKeyID:     "((ee-aws-ecr-credentials.aws_access_key_id))",
	AWSECRSecretAccessKey: "((ee-aws-ecr-credentials.aws_secret_access_key))",

	// slack
	SlackToken: "((halfpipe-slack.token))",

	// katee
	KateeKey: func(team, namespace, env string) string {
		if env == "dev" {
			return fmt.Sprintf(`((gcp:%s-%s-dev/token.token))`, team, namespace)
		}
		return fmt.Sprintf(`((gcp:%s-%s/token.token))`, team, namespace)
	},
}
