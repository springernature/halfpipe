package linters

import (
	"testing"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/parser"
	"github.com/stretchr/testify/assert"
)

const vaultPrefix = "prefix"
const foundSecret = "((found.secret))"

var calls [][]string

type StubConcourseResolv struct{}

func (StubConcourseResolv) Exists(team string, pipeline string, concourseSecret string) (err error) {
	calls = append(calls, []string{vaultPrefix, team, pipeline, concourseSecret})
	if concourseSecret == foundSecret {
		return nil
	}
	return errors.NewVaultSecretNotFoundError(vaultPrefix, team, pipeline, concourseSecret)
}

func newSecretsLinter() secretsLinter {
	calls = [][]string{}

	return secretsLinter{
		ConcourseResolv: StubConcourseResolv{},
	}
}

func TestFindSecretsDoesNothingIfThereAreNoSecrets(t *testing.T) {
	man := parser.Manifest{}
	result := newSecretsLinter().Lint(man)
	assert.Len(t, result.Errors, 0)
	assert.Len(t, calls, 0)
}

func TestErrorsForBadKeys(t *testing.T) {
	wrong1 := "((a))"
	wrong2 := "((b))"
	wrong3 := "((c))"
	man := parser.Manifest{}
	man.Team = wrong1
	man.Repo.Uri = wrong2
	man.Tasks = []parser.Task{
		parser.DeployCF{
			Password: wrong3,
		},
	}

	result := newSecretsLinter().Lint(man)
	assert.Len(t, result.Errors, 3)
	assert.Equal(t, errors.NewVaultSecretError(wrong1), result.Errors[0])
	assert.Equal(t, errors.NewVaultSecretError(wrong2), result.Errors[1])
	assert.Equal(t, errors.NewVaultSecretError(wrong3), result.Errors[2])
	assert.Len(t, calls, 0)
}

func TestReturnsErrorsIfSecretNotFound(t *testing.T) {
	notFoundSecret := "((not.found))"
	man := parser.Manifest{}
	man.Team = "team"
	man.Repo.Uri = "https://github.com/Masterminds/squirrel"
	man.Tasks = []parser.Task{
		parser.DeployCF{
			Username: foundSecret,
			Password: notFoundSecret,
		},
	}

	linter := secretsLinter{
		ConcourseResolv: StubConcourseResolv{},
	}

	result := linter.Lint(man)

	assert.Len(t, calls, 2)
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), foundSecret})
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), notFoundSecret})

	assert.Len(t, result.Errors, 1)
	assert.IsType(t, errors.VaultSecretNotFoundError{}, result.Errors[0])
}

func TestOnlyChecksForTheSameSecretOnce(t *testing.T) {
	username := "((cloudfoundry.username))"
	password := "((cloudfoundry.password))"
	api := "((cloudfoundry.api))"

	man := parser.Manifest{}
	man.Team = "team"
	man.Repo.Uri = "https://github.com/Masterminds/squirrel"
	man.Tasks = []parser.Task{
		parser.DeployCF{
			Username: foundSecret,
			Password: password,
			Api:      api,
		},
		parser.DeployCF{
			Username: username,
			Password: password,
			Api:      api,
		},
		parser.DeployCF{
			Username: username,
			Password: password,
			Api:      api,
		},
		parser.Run{
			Vars: map[string]string{
				"a": foundSecret,
				"b": password,
				"c": api,
			},
		},
		parser.DeployCF{
			Username: username,
			Password: password,
			Api:      api,
		},
	}

	result := newSecretsLinter().Lint(man)

	assert.Len(t, calls, 4)
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), foundSecret})
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), username})
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), password})
	assert.Contains(t, calls, []string{vaultPrefix, man.Team, man.Repo.GetName(), api})

	assert.Len(t, result.Errors, 3)
	assert.IsType(t, errors.VaultSecretNotFoundError{}, result.Errors[0])
}
