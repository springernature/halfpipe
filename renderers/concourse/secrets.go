package concourse

import (
	"fmt"
	"strings"
)

var vaultSecrets = struct {
	GitHubPrivateKey,
	GARToken,
	GCPArtifactsToken,
	GCRPrivateKey,
	DockerConfig,
	HalfpipeBotAppID,
	HalfpipeBotInstallationID,
	HalfpipeBotPrivateKey,
	VaultAddr,
	VaultRoleID,
	VaultSecretID,
	ConcourseURL,
	ConcoursePassword,
	ConcourseTeam,
	ConcourseUsername string
	Katee func(string) string
}{
	GitHubPrivateKey: "((halfpipe-github.private_key))",

	GARToken:          "((gcp:platform-gar/token.token))",
	GCPArtifactsToken: "((gcp:platform-artifacts/token.token))",
	GCRPrivateKey:     "((halfpipe-gcr.private_key))",
	DockerConfig:      "((halfpipe-gcr.docker_config))",

	HalfpipeBotAppID:          "((halfpipe-bot.app_id))",
	HalfpipeBotInstallationID: "((halfpipe-bot.installation_id))",
	HalfpipeBotPrivateKey:     "((halfpipe-bot.private_key))",

	VaultAddr:     "((platform/team-ro-app-role.vault_addr))",
	VaultRoleID:   "((platform/team-ro-app-role.vault_approle_id))",
	VaultSecretID: "((platform/team-ro-app-role.vault_approle_secret_id))",

	ConcourseURL:      "((platform/concourse.url))",
	ConcoursePassword: "((platform/concourse.password))",
	ConcourseTeam:     "((platform/concourse.team))",
	ConcourseUsername: "((platform/concourse.username))",

	Katee: func(namespace string) string {
		return fmt.Sprintf(`((%s-service-account-prod.key))`, strings.ReplaceAll(namespace, "katee", "katee-v2"))
	},
}
