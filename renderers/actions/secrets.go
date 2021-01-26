package actions

import (
	"fmt"
	"sort"
	"strings"
)

type Secret struct {
	vaultMap   string
	vaultField string
	outputVar  string
}

func toSecret(s string) *Secret {
	if !isSecret(s) {
		return nil
	}
	parts := strings.Split(s[2:len(s)-2], ".")
	return &Secret{
		vaultMap:   parts[0],
		vaultField: parts[1],
		outputVar:  parts[0] + "_" + parts[1],
	}
}

func isSecret(s string) bool {
	return len(strings.Split(s, ".")) == 2 && strings.HasPrefix(s, "((") && strings.HasSuffix(s, "))")
}

func (s *Secret) actionsVar() string {
	return fmt.Sprintf("${{ steps.secrets.outputs.%s }}", s.outputVar)
}

func secretsToActionsSecret(secrets []*Secret, team string) string {
	uniqueSecrets := map[string]string{}
	for _, s := range secrets {
		x := fmt.Sprintf("springernature/%s/%s %s | %s ;\n", team, s.vaultMap, s.vaultField, s.outputVar)
		uniqueSecrets[s.outputVar] = x
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
		Uses: "hashicorp/vault-action@v2.1.1",
		With: With{
			{"url", "https://vault.halfpipe.io"},
			{"method", "approle"},
			{"roleId", "${{ env.VAULT_ROLE_ID }}"},
			{"secretId", "${{ env.VAULT_SECRET_ID }}"},
			{"exportEnv", false},
			{"secrets", secretsToActionsSecret(secrets, team)},
		},
	}
}

func convertSecrets(job Job, team string) Job {
	secrets := []*Secret{}

	// job.Env
	for k, v := range job.Env {
		if s := toSecret(v); s != nil {
			secrets = append(secrets, s)
			job.Env[k] = s.actionsVar()
		}
	}

	// job.Steps.With and .Env
	newSteps := []Step{}
	for _, step := range job.Steps {
		newWith := With{}
		for _, item := range step.With {
			if s := toSecret(fmt.Sprintf("%s", item.Value)); s != nil {
				secrets = append(secrets, s)
				item.Value = s.actionsVar()
			}
			newWith = append(newWith, item)
		}

		newEnv := Env{}
		for k, v := range step.Env {
			if isSecret(v) {
				if s := toSecret(v); s != nil {
					secrets = append(secrets, s)
					newEnv[k] = s.actionsVar()
				}
			} else {
				newEnv[k] = v
			}
		}

		step.With = newWith
		step.Env = newEnv
		newSteps = append(newSteps, step)
	}
	job.Steps = newSteps

	if len(secrets) > 0 {
		job.Steps = append([]Step{fetchSecrets(secrets, team)}, job.Steps...)
	}
	return job
}
