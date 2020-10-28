package actions

func secretMapper(vaultSecret string) string {
	secrets := map[string]string{
		"((halfpipe-gcr.private_key))": "${{ secrets.GCR_PRIVATE_KEY }}",
	}
	if actionsSecret, ok := secrets[vaultSecret]; ok {
		return actionsSecret
	}
	return vaultSecret
}
