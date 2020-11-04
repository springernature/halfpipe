package actions

const repoAccessToken = "${{ secrets.EE_REPO_ACCESS_TOKEN }}"
const slackToken = "${{ secrets.EE_SLACK_TOKEN }}"

func secretMapper(vaultSecret string) string {
	secrets := map[string]string{
		"((halfpipe-gcr.private_key))": "${{ secrets.EE_GCR_PRIVATE_KEY }}",
	}
	if actionsSecret, ok := secrets[vaultSecret]; ok {
		return actionsSecret
	}
	return vaultSecret
}
