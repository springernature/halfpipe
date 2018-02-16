package linters

import (
	"github.com/springernature/halfpipe/model"
	"regexp"
	"fmt"
	"strings"
	"github.com/springernature/halfpipe/vault"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/helpers"
)

type SecretsLinter struct {
	VaultClient vault.VaultClient
	Prefix      string
}

func (s SecretsLinter) Lint(manifest model.Manifest) (result errors.LintResult) {
	result.Linter = "Secrets Linter"

	for _, secret := range s.findSecrets(manifest) {
		if s.invalidSecret(secret) {
			result.Errors = append(result.Errors, errors.NewBadVaultSecretError(secret))
		} else {
			mapName, keyName := helpers.SecretToMapAndKey(secret)
			found, err := s.VaultClient.Exists(s.Prefix, manifest.Team, manifest.Repo.GetName(), mapName, keyName)
			if err != nil {
				result.Errors = append(result.Errors, err)
			} else if !found {
				result.Errors = append(result.Errors, errors.NewNotFoundVaultSecretError(secret))
			}
		}
	}
	return
}

func (SecretsLinter) findSecrets(man model.Manifest) (secrets []string) {
	re := regexp.MustCompile(`(\(\(([^\)]+)\)\))`)
	for _, match := range re.FindAllStringSubmatch(fmt.Sprintf("%+v", man), -1) {
		secrets = append(secrets, match[1])
	}
	return
}

func (SecretsLinter) invalidSecret(secret string) bool {
	return len(strings.Split(secret, ".")) != 2
}
