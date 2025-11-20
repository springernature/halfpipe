package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe/renderers/shared/secrets"
)

var githubSecrets = struct {
	ArtifactoryUsername,
	ArtifactoryPassword,
	ArtifactoryURL,
	GCRPrivateKey,
	GitHubPrivateKey,
	RepositoryDispatchToken,
	SlackToken,
	VaultRoleID,
	VaultSecretID string
}{
	ArtifactoryUsername:     "${{ secrets.EE_ARTIFACTORY_USERNAME }}",
	ArtifactoryPassword:     "${{ secrets.EE_ARTIFACTORY_PASSWORD }}",
	ArtifactoryURL:          "${{ secrets.EE_ARTIFACTORY_URL }}",
	GCRPrivateKey:           "${{ secrets.EE_GCR_PRIVATE_KEY }}",
	GitHubPrivateKey:        "${{ secrets.EE_GITHUB_PRIVATE_KEY }}",
	RepositoryDispatchToken: "${{ secrets.EE_REPOSITORY_DISPATCH_TOKEN }}",
	SlackToken:              "${{ secrets.EE_SLACK_TOKEN }}",
	VaultRoleID:             "${{ secrets.VAULT_ROLE_ID }}",
	VaultSecretID:           "${{ secrets.VAULT_SECRET_ID }}",
}

func secretVar(s *secrets.Secret) string {
	return fmt.Sprintf("${{ steps.secrets.outputs.%s }}", secretOutputVar(s))
}

func secretVaultPath(s *secrets.Secret) string {
	return fmt.Sprintf("/springernature/data/%s %s", s.MapPath, s.Key)
}
func secretOutputVar(s *secrets.Secret) string {
	ov := strings.ReplaceAll(secretVaultPath(s), "/", "_")
	ov = strings.ReplaceAll(ov, " ", "_")
	ov = strings.TrimPrefix(ov, "_")
	return ov
}

func secretsToActionsSecret(secrets []*secrets.Secret) string {
	uniqueSecrets := map[string]string{}
	for _, s := range secrets {
		x := fmt.Sprintf("%s | %s ;\n", secretVaultPath(s), secretOutputVar(s))
		uniqueSecrets[secretOutputVar(s)] = x
	}

	var secs []string
	for _, v := range uniqueSecrets {
		secs = append(secs, v)
	}
	sort.Strings(secs)

	return strings.Join(secs, "")
}

func fetchSecrets(secrets []*secrets.Secret) Step {
	return Step{
		Name: "Vault secrets",
		ID:   "secrets",
		Uses: ExternalActions.Vault,
		With: With{
			"url":       "https://vault.halfpipe.io",
			"method":    "approle",
			"roleId":    "${{ env.VAULT_ROLE_ID }}",
			"secretId":  "${{ env.VAULT_SECRET_ID }}",
			"exportEnv": false,
			"secrets":   secretsToActionsSecret(secrets),
		},
	}
}

func convertSecrets(steps Steps, team string) (newSteps Steps) {
	allSecrets := []*secrets.Secret{}

	for _, step := range steps {
		newWith := With{}
		for key, value := range step.With {
			switch v := value.(type) {
			case MultiLine:
				secretList, multiLineStringWithActionSecret := multiLineStringToSecret(v.m, team)
				allSecrets = append(allSecrets, secretList...)
				value = MultiLine{multiLineStringWithActionSecret}
			default:
				if s := secrets.New(fmt.Sprintf("%v", value), team); s != nil {
					allSecrets = append(allSecrets, s)
					value = secretVar(s)
				}
			}
			newWith[key] = value
		}
		step.With = newWith
		for k, v := range step.Env {
			if s := secrets.New(v, team); s != nil {
				allSecrets = append(allSecrets, s)
				step.Env[k] = secretVar(s)
			}
		}
		newSteps = append(newSteps, step)
	}

	if len(allSecrets) > 0 {
		newSteps = append(Steps{fetchSecrets(allSecrets)}, newSteps...)
	}
	return newSteps
}

func multiLineStringToSecret(ml map[string]string, team string) ([]*secrets.Secret, map[string]string) {
	m := make(map[string]string)
	var sec []*secrets.Secret
	for k, v := range ml {
		if a := secrets.New(v, team); a != nil {
			sec = append(sec, a)
			m[k] = secretVar(a)
		} else {
			m[k] = v
		}
	}
	return sec, m
}
