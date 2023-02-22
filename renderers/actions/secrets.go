package actions

import (
	"fmt"
	"sort"
	"strings"
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

type Secret struct {
	vaultPath string
}

func (s *Secret) actionsVar() string {
	return fmt.Sprintf("${{ steps.secrets.outputs.%s }}", s.outputVar())
}

func (s *Secret) outputVar() string {
	ov := strings.ReplaceAll(s.vaultPath, "/", "_")
	ov = strings.ReplaceAll(ov, " ", "_")
	ov = strings.TrimPrefix(ov, "_")
	return ov
}

func isShared(s string) bool {
	return map[string]bool{
		"PPG-coco-gateway":   true,
		"artifactory":        true,
		"artifactory_test":   true,
		"contrastsecurity":   true,
		"grafana":            true,
		"halfpipe-artifacts": true,
		"halfpipe-gcr":       true,
		"halfpipe-github":    true,
		"halfpipe-ml-deploy": true,
		"halfpipe-semver":    true,
		"halfpipe-slack":     true,
	}[s]
}

func toSecret(s string, team string) *Secret {

	if !isSecret(s) {
		return nil
	}

	secretValue := s[2 : len(s)-2]

	if isKeyValueSecret(secretValue) {
		parts := strings.Split(secretValue, ".")
		if isShared(parts[0]) {
			team = "shared"
		}
		return &Secret{
			vaultPath: fmt.Sprintf("/springernature/data/%s/%s %s", team, parts[0], parts[1]),
		}
	}

	if isAbsolutePathSecret(secretValue) {
		return &Secret{
			vaultPath: secretValue,
		}
	}

	return nil
}

func isSecret(s string) bool {
	return strings.HasPrefix(s, "((") && strings.HasSuffix(s, "))")
}

func isAbsolutePathSecret(s string) bool {
	return len(strings.Split(s, " ")) == 2
}

func isKeyValueSecret(s string) bool {
	return len(strings.Split(s, ".")) == 2
}

func secretsToActionsSecret(secrets []*Secret) string {
	uniqueSecrets := map[string]string{}
	for _, s := range secrets {
		x := fmt.Sprintf("%s | %s ;\n", s.vaultPath, s.outputVar())
		uniqueSecrets[s.outputVar()] = x
	}

	var secs []string
	for _, v := range uniqueSecrets {
		secs = append(secs, v)
	}
	sort.Strings(secs)

	return strings.Join(secs, "")
}

func fetchSecrets(secrets []*Secret, team string) Step {
	return Step{
		Name: "Vault secrets",
		ID:   "secrets",
		Uses: "hashicorp/vault-action@v2.4.3",
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
	secrets := []*Secret{}

	for _, step := range steps {
		newWith := With{}
		for key, value := range step.With {
			multiLineStringArray := strings.Split(fmt.Sprintf("%s", value), "\n")
			if s := toSecret(fmt.Sprintf("%s", value), team); s != nil {
				secrets = append(secrets, s)
				value = s.actionsVar()
			} else if len(multiLineStringArray) > 1 {
				secretList, multiLineStringWithActionSecret := multiLineStringToSecret(multiLineStringArray, team)
				secrets = append(secrets, secretList...)
				value = multiLineStringWithActionSecret
			}
			newWith[key] = value
		}
		step.With = newWith
		for k, v := range step.Env {
			if s := toSecret(v, team); s != nil {
				secrets = append(secrets, s)
				step.Env[k] = s.actionsVar()
			}
		}
		newSteps = append(newSteps, step)
	}

	if len(secrets) > 0 {
		newSteps = append(Steps{fetchSecrets(secrets, team)}, newSteps...)
	}
	return newSteps
}

func multiLineStringToSecret(multiLineStringArray []string, team string) (sec []*Secret, newBuildArgs string) {
	var newBuildArgsArray []string
	for _, line := range multiLineStringArray {
		keyValueArray := strings.Split(line, "=")
		if len(keyValueArray) > 1 {
			if a := toSecret(keyValueArray[1], team); a != nil {
				sec = append(sec, a)
				newBuildArgsArray = append(newBuildArgsArray, fmt.Sprintf("%s=%s", keyValueArray[0], a.actionsVar()))
			} else {
				newBuildArgsArray = append(newBuildArgsArray, strings.Join(keyValueArray, "="))
			}
		} else {
			newBuildArgsArray = append(newBuildArgsArray, keyValueArray[0])
		}
	}
	return sec, strings.Join(newBuildArgsArray, "\n")
}
