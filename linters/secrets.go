package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

type secretsLinter struct{}

func NewSecretsLinter() Linter {
	return secretsLinter{}
}

func (s secretsLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Secrets"
	result.DocsURL = "https://docs.halfpipe.io/vault/"

	for _, sec := range findSecrets(manifest) {
		if !secretIsValidFormat(sec) || !secretHasOnlyValidChars(sec) {
			result.AddError(errors.NewVaultSecretError(sec))
		}
	}
	return
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

func secretHasOnlyValidChars(secret string) bool {
	return regexp.MustCompile(`^\(\([a-zA-Z0-9\-_\.]+\)\)$`).MatchString(secret)
}
