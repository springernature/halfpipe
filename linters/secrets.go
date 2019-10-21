package linters

import (
	"github.com/springernature/halfpipe/linters/result"
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

func (s secretsLinter) Lint(manifest manifest.Manifest) (result result.LintResult) {
	result.Linter = "Secrets"
	result.DocsURL = "https://docs.halfpipe.io/vault/"

	result.AddError(s.secretValidator.Validate(manifest)...)
	return result
}
