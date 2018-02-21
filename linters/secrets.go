package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/helpers"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/vault"
)

type SecretsLinter struct {
	VaultClient vault.Client
}

func (s SecretsLinter) Lint(manifest model.Manifest) (result model.LintResult) {
	result.Linter = "Secrets Linter"

	for _, secret := range s.findSecrets(manifest) {
		if s.invalidSecret(secret) {
			result.Errors = append(result.Errors, errors.NewVaultSecretError(secret))
		} else {
			mapName, keyName := helpers.SecretToMapAndKey(secret)
			team := manifest.Team
			pipeline := manifest.Repo.GetName()
			found, err := s.VaultClient.Exists(team, pipeline, mapName, keyName)
			if err != nil {
				result.Errors = append(result.Errors, err)
			} else if !found {
				result.Errors = append(result.Errors, errors.NewVaultSecretNotFoundError(s.VaultClient.VaultPrefix(), team, pipeline, secret))
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
