package config

const HalfpipeBotName = "halfpipe-bot"

var GitHubSecrets = struct {
	ArtifactoryUsername, ArtifactoryPassword, ArtifactoryURL,
	HalfpipeBotClientID, HalfpipeBotPrivateKey,
	SlackToken, VaultAddr, VaultRoleID, VaultSecretID string
}{
	ArtifactoryUsername:   "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
	ArtifactoryPassword:   "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	ArtifactoryURL:        "${{ secrets.EE_ARTIFACTORY_URL }}",
	HalfpipeBotClientID:   "${{ secrets.EE_HALFPIPE_BOT_CLIENT_ID }}",
	HalfpipeBotPrivateKey: "${{ secrets.EE_HALFPIPE_BOT_PRIVATE_KEY }}",
	SlackToken:            "${{ secrets.EE_SLACK_TOKEN }}",
	VaultAddr:             "${{ secrets.EE_VAULT_ADDR }}",
	VaultRoleID:           "${{ secrets.VAULT_ROLE_ID }}",
	VaultSecretID:         "${{ secrets.VAULT_SECRET_ID }}",
}
