package linter

import (
	"regexp"

	"fmt"

	. "github.com/robwhitby/halfpipe-cli/model"
)

func requiredSecrets(man Manifest) (secrets []string) {
	re := regexp.MustCompile(`\(\(([^\)]+)\)\)`)
	for _, match := range re.FindAllStringSubmatch(fmt.Sprintf("%+v", man), -1) {
		secrets = append(secrets, match[1])
	}
	return
}

func LintSecrets(man Manifest, secretChecker SecretChecker) (errors []error) {
	for _, secret := range requiredSecrets(man) {
		if !secretChecker(secret) {
			errors = append(errors, NewMissingSecret(secret))
		}
	}
	return
}
