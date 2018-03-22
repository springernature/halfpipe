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
	result.DocsURL = "https://docs.halfpipe.io/docs/vault/"

	if manifest.Team == "" {
		return
	}

	allSecrets := findSecrets(manifest)
	if len(allSecrets) == 0 {
		return
	}

	store, err := s.secretStoreFunc()
	if err != nil {
		result.AddWarning(err)
		return
	}

	chSecretErrs := make(chan error)

	for _, sec := range allSecrets {
		go func(str string) {
			chSecretErrs <- s.checkExists(store, manifest.Team, manifest.Repo.GetName(), str)
		}(sec)
	}
	for range allSecrets {
		err := <-chSecretErrs
		if err != nil {
			result.AddWarning(err)
		}
	}

	return
}

func (s secretsLinter) checkExists(store secrets.SecretStore, team string, pipeline string, concourseSecret string) error {
	if !secretIsValidFormat(concourseSecret) {
		return errors.NewVaultSecretError(concourseSecret)
	}

	secretMap, secretKey := secretToMapAndKey(concourseSecret)
	paths := []string{
		path.Join(s.prefix, team, pipeline, secretMap),
		path.Join(s.prefix, team, secretMap),
	}

	for _, p := range paths {
		exists, err := store.Exists(p, secretKey)
		if exists || err != nil {
			return err
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
