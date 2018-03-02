package linters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/linters/secret_resolver"
	"github.com/springernature/halfpipe/model"
)

type secretsLinter struct {
	ConcourseResolv secret_resolver.ConcourseResolver
}

func (s secretsLinter) Lint(manifest model.Manifest) (result model.LintResult) {
	result.Linter = "Secrets"
	if manifest.Team == "" {
		return
	}

	for _, secret := range s.findSecrets(manifest) {
		if s.invalidSecret(secret) {
			result.Errors = append(result.Errors, errors.NewVaultSecretError(secret))
		} else {
			if err := s.ConcourseResolv.Exists(manifest.Team, manifest.Repo.GetName(), secret); err != nil {
				result.AddError(err)
			}
		}
	}
	return
}

func NewSecretsLinter(resolver secret_resolver.ConcourseResolver) Linter {
	return secretsLinter{
		ConcourseResolv: resolver,
	}
}

func (secretsLinter) findSecrets(man model.Manifest) (secrets []string) {
	re := regexp.MustCompile(`(\(\(([^\)]+)\)\))`)
	for _, match := range re.FindAllStringSubmatch(fmt.Sprintf("%+v", man), -1) {
		if !secretAlreadySeen(match[1], secrets) {
			secrets = append(secrets, match[1])
		}
	}
	return
}

func (secretsLinter) invalidSecret(secret string) bool {
	return len(strings.Split(secret, ".")) != 2
}

func secretAlreadySeen(secret string, secrets []string) bool {
	// This is stupid. But people will not have thousands of secrets so fuck it

	for _, s := range secrets {
		if s == secret {
			return true
		}
	}
	return false
}
