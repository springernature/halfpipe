package linters

import (
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FakeVaultClient struct {
}

var lastCalledKey = ""

func (FakeVaultClient) Exists(path string) (bool, error) {
	lastCalledKey = path
	return true, nil
}

var secretsLinter = SecretsLinter{FakeVaultClient{}}

//func TestFindSecretsPlaceholder(t *testing.T) {
//	man := model.Manifest{}
//	man.Tasks = []model.Task{
//		model.DeployCF{
//			Password: "((supersecret.password))",
//		},
//	}
//
//	result := secretsLinter.Lint(man)
//	assert.Len(t, result.Errors, 1)
//	assertMissingField(t, "team", result.Errors[0])
//}

func TestFindSecretsDoesNothingIfThereAreNoSecrets(t *testing.T) {
	man := model.Manifest{}
	result := secretsLinter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestFindSecretsPlaceholder(t *testing.T) {
	wrong1 := "((a))"
	wrong2 := "((b))"
	wrong3 := "((c))"
	man := model.Manifest{}
	man.Team = wrong1
	man.Repo.Uri = wrong2
	man.Tasks = []model.Task{
		model.DeployCF{
			Password: wrong3,
		},
	}

	result := secretsLinter.Lint(man)
	assert.Len(t, result.Errors, 3)
	assertVaultError(t, "a", result.Errors[0])
	assertVaultError(t, "b", result.Errors[1])
	assertVaultError(t, "c", result.Errors[2])
}

func TestFindSecretsReturnsErrorIfASecretIsMalformed(t *testing.T) {
	key := "((a))"
	man := model.Manifest{
		Repo: model.Repo{
			Uri:        "((my.uri))",
			PrivateKey: key,
		},
	}
	result := secretsLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertVaultError(t, "a", result.Errors[0])
}

func TestFindSecretsCallsOutToVaultClientForValidKeys(t *testing.T) {
	key := "((a))"
	man := model.Manifest{
		Team: key,
		Repo: model.Repo{
			Uri: "((my.uri))",
		},
	}
	result := secretsLinter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertVaultError(t, "a", result.Errors[0])
	assert.Equal(t, "my.uri", lastCalledKey)

}

// Given this manifest
// I call out in this way
// And if stuff returns from dep
// I give back these errors.
