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

// check if a secret matches one of the shared secrets
// vault kv list /springernature/shared
func isShared(s string) bool {
	return map[string]bool{
		"PPG-gradle-version-reporter":   true,
		"PPG-owasp-dependency-reporter": true,
		"artifactory":                   true,
		"artifactory-support":           true,
		"artifactory_test":              true,
		"bla":                           true,
		"burpsuiteenterprise":           true,
		"contrastsecurity":              true,
		"ee-sso-route-service":          true,
		"fastly":                        true,
		"grafana":                       true,
		"halfpipe-artifacts":            true,
		"halfpipe-docker-config":        true,
		"halfpipe-gcr":                  true,
		"halfpipe-github":               true,
		"halfpipe-ml-deploy":            true,
		"halfpipe-semver":               true,
		"halfpipe-slack":                true,
		"katee-tls-dev":                 true,
		"katee-tls-prod":                true,
		"sentry-release-integration":    true,
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
		Uses: "hashicorp/vault-action@v2.5.0",
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
			switch v := value.(type) {
			case MultiLine:
				secretList, multiLineStringWithActionSecret := multiLineStringToSecret(v.m, team)
				secrets = append(secrets, secretList...)
				value = MultiLine{multiLineStringWithActionSecret}
			default:
				if s := toSecret(fmt.Sprintf("%v", value), team); s != nil {
					secrets = append(secrets, s)
					value = s.actionsVar()
				}
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

func multiLineStringToSecret(ml map[string]string, team string) ([]*Secret, map[string]string) {
	m := make(map[string]string)
	var sec []*Secret
	for k, v := range ml {
		if a := toSecret(v, team); a != nil {
			sec = append(sec, a)
			m[k] = a.actionsVar()
		} else {
			m[k] = v
		}
	}
	return sec, m
}
