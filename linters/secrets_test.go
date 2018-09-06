package linters

import (
	"testing"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

func TestNoSecrets(t *testing.T) {
	linter := NewSecretsLinter()
	man := manifest.Manifest{}

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 0)
}

func TestBadKeys(t *testing.T) {
	linter := NewSecretsLinter()

	wrong1 := "((a))"
	wrong2 := "((b))"
	wrong3 := "((c.d.e))"
	man := manifest.Manifest{}
	man.Team = wrong1
	man.Repo.URI = wrong2
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Password: wrong3,
		},
	}

	result := linter.Lint(man)
	if assert.Len(t, result.Errors, 3) {
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong1))
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong2))
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong3))
	}
}

func TestBadCharacters(t *testing.T) {
	linter := NewSecretsLinter()

	wrong1 := "((this_is_a_invalid$secret.@with_special_chars))"
	wrong2 := "((this_is_a_invalid%secret.*with_special_chars))"
	wrong3 := "((this_is_a_invalid#secret.+with_special_chars))"
	man := manifest.Manifest{}
	man.Team = wrong1
	man.Repo.URI = wrong2
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Password: wrong3,
		},
	}

	result := linter.Lint(man)
	if assert.Len(t, result.Errors, 3) {
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong1))
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong2))
		assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong3))
	}
}
