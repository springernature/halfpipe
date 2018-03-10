package linters

import (
	"testing"

	"path"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/linters/secrets"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
)

const (
	vaultPrefix     = "prefix"
	foundSecret     = "((found.secret))"
	prefix          = "prefix"
	team            = "team"
	pipeline        = "pipeline"
	mapKey          = "mapKey"
	secretKey       = "secretKey"
	concourseSecret = "((" + mapKey + "." + secretKey + "))"
)

type fakeSecretStore struct {
	Calls      [][]string
	ExistsFunc func(path, secretKey string) (bool, error)
}

func (s *fakeSecretStore) Exists(path string, secretKey string) (exists bool, err error) {
	s.Calls = append(s.Calls, []string{path, secretKey})
	return s.ExistsFunc(path, secretKey)
}

func secretsLinterWithFakeStore(exists bool, err error) (secretsLinter, *fakeSecretStore) {
	store := &fakeSecretStore{
		ExistsFunc: func(path, secretKey string) (bool, error) {
			return exists, err
		},
	}
	storeFunc := func() (secrets.SecretStore, error) { return store, nil }
	return NewSecretsLinter(vaultPrefix, storeFunc), store
}

func TestNoSecrets(t *testing.T) {
	linter, _ := secretsLinterWithFakeStore(false, nil)
	man := manifest.Manifest{}

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)
}

func TestBadKeys(t *testing.T) {
	linter, _ := secretsLinterWithFakeStore(false, nil)
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
	assert.Len(t, result.Errors, 3)
	assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong1))
	assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong2))
	assert.Contains(t, result.Errors, errors.NewVaultSecretError(wrong3))
}

func TestSecretNotFound(t *testing.T) {
	notFoundSecret := "((not.found))"
	man := manifest.Manifest{}
	man.Team = "team"
	man.Repo.URI = "https://github.com/Masterminds/squirrel"
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Username: "user",
			Password: notFoundSecret,
		},
	}

	linter, _ := secretsLinterWithFakeStore(false, nil)
	result := linter.Lint(man)

	assert.Len(t, result.Errors, 1)
	assert.IsType(t, errors.VaultSecretNotFoundError{}, result.Errors[0])
}

func TestSecretFound(t *testing.T) {
	man := manifest.Manifest{}
	man.Team = "team"
	man.Repo.URI = "https://github.com/Masterminds/squirrel"
	man.Tasks = []manifest.Task{
		manifest.DeployCF{
			Username: "user",
			Password: foundSecret,
		},
	}

	linter, _ := secretsLinterWithFakeStore(true, nil)
	result := linter.Lint(man)

	assert.Len(t, result.Errors, 0)

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

	secrets := findSecrets(man)
	assert.Len(t, secrets, 2)
}

func TestCallsOutToTwoDifferentPathsAndReturnsErrorIfNotFound(t *testing.T) {
	linter, store := secretsLinterWithFakeStore(false, nil)

	err := linter.checkExists(store, team, pipeline, concourseSecret)

	assert.Len(t, store.Calls, 2)
	assert.Contains(t, store.Calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Contains(t, store.Calls, []string{path.Join(prefix, team, mapKey), secretKey})

	assert.NotNil(t, err)
}

func TestCallsOutOnlyOnceIfWeFindTheSecretInTheFirstPath(t *testing.T) {
	linter, store := secretsLinterWithFakeStore(true, nil)

	err := linter.checkExists(store, team, pipeline, concourseSecret)

	assert.Len(t, store.Calls, 1)
	assert.Contains(t, store.Calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Nil(t, err)
}

func TestCallsOutTwiceAndReturnsNilIfFoundInSecondCall(t *testing.T) {
	linter, store := secretsLinterWithFakeStore(true, nil)

	store.ExistsFunc = func(path string, secretKey string) (bool, error) {
		return len(store.Calls) == 2, nil
	}

	err := linter.checkExists(store, team, pipeline, concourseSecret)

	assert.Len(t, store.Calls, 2)
	assert.Contains(t, store.Calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Contains(t, store.Calls, []string{path.Join(prefix, team, mapKey), secretKey})
	assert.Nil(t, err)

}

func TestRaisesErrorFromInitialisingStoreWhenThereAreSecrets(t *testing.T) {
	myError := errors.NewVaultSecretError("whatever")
	store := &fakeSecretStore{
		ExistsFunc: func(path, secretKey string) (bool, error) {
			return false, myError
		},
	}
	storeFunc := func() (secrets.SecretStore, error) { return store, nil }

	linter := NewSecretsLinter(vaultPrefix, storeFunc)

	withoutSecretsResult := linter.Lint(manifest.Manifest{})
	assert.False(t, withoutSecretsResult.HasErrors())

	withSecretsResult := linter.Lint(manifest.Manifest{Team: concourseSecret})
	assert.Len(t, withSecretsResult.Errors, 1)
	assert.Equal(t, myError, withSecretsResult.Errors[0])
}

func TestRaisesErrorFromStore(t *testing.T) {
	myError := errors.NewVaultSecretError("whatever")
	linter, store := secretsLinterWithFakeStore(false, myError)

	err := linter.checkExists(store, team, pipeline, concourseSecret)
	assert.Len(t, store.Calls, 1)
	assert.Equal(t, myError, err)
}
