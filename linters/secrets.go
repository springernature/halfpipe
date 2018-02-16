package linters

import (
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/errors"
	"regexp"
	"fmt"
	"strings"
	"github.com/springernature/halfpipe/vault"
)

var VaultPathWithRepoName = "/springernature/%s/%s/%s %s"
var VaultPathWithoutRepoName = "/springernature/%s/%s %s"

type SecretsLinter struct {
	VaultClient vault.VaultClient
}

func (secretsLinter SecretsLinter) Lint(man model.Manifest) (result errors.LintResult) {
	result.Linter = "Secrets Linter"

	for _, secret := range requiredSecrets(man) {
		if !validKey(secret) {
			result.Errors = append(result.Errors, errors.NewVaultError(secret, ""))
		} else {
			parts := strings.Split(secret, ".")
			found, err := secretsLinter.secretExists(man, fmt.Sprintf(VaultPathWithRepoName, man.Team, man.Repo.GetName(), parts[0], parts[1]))
			if err != nil {
				result.Errors = append(result.Errors, err)
				break
			} else if found {
				break
			}

			found, err = secretsLinter.secretExists(man, fmt.Sprintf(VaultPathWithoutRepoName, man.Team, parts[0], parts[1]))
			if err != nil {
				result.Errors = append(result.Errors, err)
				break
			}
			if !found {
				result.Errors = append(result.Errors, errors.NewVaultError(secret, ""))
			}
		}
	}
	return
}

func (secretsLinter SecretsLinter) secretExists(manifest model.Manifest, key string) (bool, error) {
	exists, err := secretsLinter.VaultClient.Exists(key)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}
	return true, nil
}

func requiredSecrets(man model.Manifest) (secrets []string) {
	re := regexp.MustCompile(`\(\(([^\)]+)\)\)`)
	for _, match := range re.FindAllStringSubmatch(fmt.Sprintf("%+v", man), -1) {
		secrets = append(secrets, match[1])
	}
	return
}

func validKey(key string) bool {
	return len(strings.Split(key, ".")) == 2
}

/*

func buildVaultPath(man model.Manifest, secret string) ([]string) {
	return []string{
		fmt.Sprintf("/springernature/%s/%s", man.Team, secret),
		fmt.Sprintf("/springernature/%s/%/%s", man.Team, man.Repo.GetName(), secret),
		}
}


func secretChecker(path string) (error) {

	config := api.DefaultConfig()
	if config.Address == "https://127.0.0.1:8200" {
		return errors.NewVaultError("Missing VAULT_ADDR")
	}

	client, err := api.NewClient(config)
	if err != nil {
		// bla bla errror
	}

	var mysec *api.Secret
	var error error

	mysec, error = client.Logical().Read(path)
	if error != nil {
		fmt.Print(error)
	} else if mysec == nil {
		fmt.Print("No secret found")
	}

	return true
}
*/
