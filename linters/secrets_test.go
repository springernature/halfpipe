package linters

import (
	"testing"

	"path"
	"strings"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/secrets"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

const (
	prefix   = "prefix"
	team     = "team"
	pipeline = "pipeline"
)

type fakeSecretStore struct {
	ExistsFunc func(path, secretKey string) (bool, error)
}

func NewFakeSecretStore(existsFunc func(path, secretKey string) (bool, error)) secrets.SecretStore {
	return fakeSecretStore{
		ExistsFunc: existsFunc,
	}
}

func (s fakeSecretStore) Exists(path string, secretKey string) (exists bool, err error) {
	return s.ExistsFunc(path, secretKey)
}

func TestNoSecrets(t *testing.T) {
	linter := NewSecretsLinter(prefix, func() (ss secrets.SecretStore, err error) { return })
	man := manifest.Manifest{}

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 0)
}

func TestBadKeys(t *testing.T) {
	linter := NewSecretsLinter(prefix, func() (ss secrets.SecretStore, err error) { return })

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
	if assert.Len(t, result.Warnings, 3) {
		assert.Contains(t, result.Warnings, errors.NewVaultSecretError(wrong1))
		assert.Contains(t, result.Warnings, errors.NewVaultSecretError(wrong2))
		assert.Contains(t, result.Warnings, errors.NewVaultSecretError(wrong3))
	}
}

func TestSecretNotFound(t *testing.T) {
	notFoundSecret := "((not.found))"
	man := manifest.Manifest{}
	man.Team = team
	man.Pipeline = pipeline
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Username: "user",
			Password: notFoundSecret,
		},
	}

	var paths []string
	var secretKey string
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			paths = append(paths, path)
			secretKey = sK
			return false, nil
		}), nil
	})

	result := linter.Lint(man)

	assert.Len(t, result.Warnings, 1)
	assert.IsType(t, errors.VaultSecretNotFoundError{}, result.Warnings[0])
	assert.Equal(t, path.Join(prefix, team, pipeline, "not"), paths[0])
	assert.Equal(t, path.Join(prefix, team, "not"), paths[1])
	assert.Equal(t, "found", secretKey)
}

func TestCallsOnlyOutOnceIfFoundInFirstPath(t *testing.T) {
	man := manifest.Manifest{
		Team:     team,
		Pipeline: pipeline,
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Username: "user",
				Password: "((my.secret))",
			},
		},
	}

	var paths []string
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			paths = append(paths, path)
			return true, nil
		}), nil
	})

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, paths, 1)
	assert.Equal(t, path.Join(prefix, team, pipeline, "my"), paths[0])
}

func TestCallsOnlyTwiceToFindSecret(t *testing.T) {
	man := manifest.Manifest{
		Team:     team,
		Pipeline: pipeline,
		Tasks: []manifest.Task{
			manifest.DeployCF{
				Username: "user",
				Password: "((my.secret))",
			},
		},
	}

	var paths []string
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			paths = append(paths, path)
			if strings.Contains(path, pipeline) {
				return false, nil
			}
			return true, nil
		}), nil
	})

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)
	assert.Len(t, result.Warnings, 0)
	assert.Len(t, paths, 2)
}

func TestOnlyChecksForTheSameSecretOnce(t *testing.T) {
	username := "((cloudfoundry.username))"
	password := "((cloudfoundry.password))"

	man := manifest.Manifest{}
	man.Team = "team"
	man.Tasks = []manifest.Task{
		manifest.Run{
			Vars: map[string]string{
				"a": username,
				"b": username,
				"c": password,
			},
		},
	}

	var numCalls int
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			numCalls++
			return true, nil
		}), nil
	})

	linter.Lint(man)
	assert.Equal(t, numCalls, 2)
}

func TestRaisesWarningFromInitialisingStoreWhenThereAreSecrets(t *testing.T) {
	myError := errors.NewVaultSecretError("whatever")
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			return false, myError
		}), nil
	})

	withoutSecretsResult := linter.Lint(manifest.Manifest{})
	assert.False(t, withoutSecretsResult.HasErrors())

	withSecretsResult := linter.Lint(manifest.Manifest{Team: "((teams.team))"})
	assert.Len(t, withSecretsResult.Warnings, 1)
	assert.Equal(t, myError, withSecretsResult.Warnings[0])
}

func TestRaisesErrorFromStore(t *testing.T) {
	myError := errors.NewVaultSecretError("whatever")
	linter := NewSecretsLinter(prefix, func() (secrets.SecretStore, error) {
		return NewFakeSecretStore(func(path, sK string) (bool, error) {
			return false, nil
		}), myError
	})

	result := linter.Lint(manifest.Manifest{Team: "MyTeam", Repo: manifest.Repo{URI: "((a.secret))"}})
	assert.Equal(t, myError, result.Warnings[0])
}
