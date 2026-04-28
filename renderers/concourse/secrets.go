package concourse

import (
	"fmt"
	"strings"
)

var secrets = struct {
	GitHubPrivateKey, GithubStatusesToken, HalfpipeBotAppID, HalfpipeBotInstallationID, HalfpipeBotPrivateKey,
	GARToken, GCPArtifactsToken, GCRPrivateKey, VersionBucket, VersionJSONKey, ArtifactsBucket,
	VaultAddr, VaultRoleID, VaultSecretID,
	ConcourseURL, ConcoursePassword, ConcourseTeam, ConcourseUsername,
	SlackToken string
	KateeKey func(string) string
}{
	// github
	GitHubPrivateKey:          "((halfpipe-github.private_key))",
	GithubStatusesToken:       "((halfpipe-github.statuses-token))",
	HalfpipeBotAppID:          "((halfpipe-bot.app_id))",
	HalfpipeBotInstallationID: "((halfpipe-bot.installation_id))",
	HalfpipeBotPrivateKey:     "((halfpipe-bot.private_key))",

	// gcp
	GARToken:          "((gcp:platform-gar/token.token))",
	GCPArtifactsToken: "((gcp:platform-artifacts/token.token))",
	GCRPrivateKey:     "((halfpipe-gcr.private_key))",
	VersionBucket:     "((halfpipe-semver.bucket))",
	VersionJSONKey:    "((halfpipe-semver.private_key))",
	ArtifactsBucket:   "((halfpipe-artifacts.bucket))",

	// vault
	VaultAddr:     "((platform/team-ro-app-role.vault_addr))",
	VaultRoleID:   "((platform/team-ro-app-role.vault_approle_id))",
	VaultSecretID: "((platform/team-ro-app-role.vault_approle_secret_id))",

	// concourse
	ConcourseURL:      "((platform/concourse.url))",
	ConcoursePassword: "((platform/concourse.password))",
	ConcourseTeam:     "((platform/concourse.team))",
	ConcourseUsername: "((platform/concourse.username))",

	// slack
	SlackToken: "((halfpipe-slack.token))",

	// katee
	KateeKey: func(namespace string) string {
		return fmt.Sprintf(`((%s-service-account-prod.key))`, strings.ReplaceAll(namespace, "katee", "katee-v2"))
	},
}
