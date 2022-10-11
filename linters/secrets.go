package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

type secretsLinter struct {
	secretValidator manifest.SecretValidator
}

func NewSecretsLinter(secretValidator manifest.SecretValidator) Linter {
	return secretsLinter{
		secretValidator: secretValidator,
	}
}

func (s secretsLinter) Lint(manifest manifest.Manifest) (result LintResult) {
	result.Linter = "Secrets"
	result.DocsURL = "https://ee.public.springernature.app/rel-eng/vault/"

	result.Add(s.secretValidator.Validate(manifest)...)
	return result
}
