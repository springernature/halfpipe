package linters

import (
	"fmt"
	"regexp"
	"strings"

	"path"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/secrets"
	"github.com/springernature/halfpipe/manifest"
)

type secretsLinter struct {
	prefix          string
	secretStoreFunc secrets.SecretStoreFunc
}

func NewSecretsLinter(secretPrefix string, storeFunc secrets.SecretStoreFunc) secretsLinter {
	return secretsLinter{
		prefix:          secretPrefix,
		secretStoreFunc: storeFunc,
	}
}

func (s secretsLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Secrets"
	if manifest.Team == "" {
		return
	}

	secrets := findSecrets(manifest)
	if len(secrets) == 0 {
		return
	}

	store, err := s.secretStoreFunc()
	if err != nil {
		result.Errors = append(result.Errors, err)
		return
	}

	for _, secret := range secrets {
		if !secretIsValidFormat(secret) {
			result.Errors = append(result.Errors, errors.NewVaultSecretError(secret))
		} else {
			if err := s.checkExists(store, manifest.Team, manifest.Repo.GetName(), secret); err != nil {
				result.AddError(err)
			}
		}
	}
	return
}

func (s secretsLinter) checkExists(store secrets.SecretStore, team string, pipeline string, concourseSecret string) (err error) {
	secretMap, secretKey := secretToMapAndKey(concourseSecret)
	paths := []string{
		path.Join(s.prefix, team, pipeline, secretMap),
		path.Join(s.prefix, team, secretMap),
	}

	for _, p := range paths {
		exists, e := store.Exists(p, secretKey)

		if exists || e != nil {
			err = e
			return
		}
	}

	return errors.NewVaultSecretNotFoundError(s.prefix, team, pipeline, concourseSecret)
}

func findSecrets(man manifest.Manifest) (secrets []string) {
	re := regexp.MustCompile(`(\(\(([^\)]+)\)\))`)
	set := make(map[string]bool)
	for _, match := range re.FindAllStringSubmatch(fmt.Sprintf("%+v", man), -1) {
		set[match[1]] = true
	}
	for key := range set {
		secrets = append(secrets, key)
	}
	return
}

func secretIsValidFormat(secret string) bool {
	return len(strings.Split(secret, ".")) == 2
}

func secretToMapAndKey(secret string) (string, string) {
	s := strings.Replace(strings.Replace(secret, "((", "", -1), "))", "", -1)
	parts := strings.Split(s, ".")
	mapName, keyName := parts[0], parts[1]
	return mapName, keyName
}
