package linters

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/springernature/halfpipe/errors"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/vault"
	"github.com/stretchr/testify/assert"
)

func TestFindSecretsDoesNothingIfThereAreNoSecrets(t *testing.T) {
	man := model.Manifest{}
	result := SecretsLinter{}.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestErrorsForBadKeys(t *testing.T) {
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

	result := SecretsLinter{}.Lint(man)
	assert.Len(t, result.Errors, 3)
	assert.Equal(t, errors.NewBadVaultSecretError(wrong1), result.Errors[0])
	assert.Equal(t, errors.NewBadVaultSecretError(wrong2), result.Errors[1])
	assert.Equal(t, errors.NewBadVaultSecretError(wrong3), result.Errors[2])
}

func TestReturnsErrorsIfSecretNotFound(t *testing.T) {
	foundSecret := "((found.secret))"
	notFoundSecret := "((not.found))"
	man := model.Manifest{}
	man.Team = "team"
	man.Repo.Uri = "https://github.com/Masterminds/squirrel"
	man.Tasks = []model.Task{
		model.DeployCF{
			Username: foundSecret,
			Password: notFoundSecret,
		},
	}

	ctrl := gomock.NewController(t)
	mockClient := vault.NewMockClient(ctrl)
	linter := SecretsLinter{
		mockClient,
	}

	pipelineName := man.Repo.GetName()
	mockClient.EXPECT().Exists(man.Team, pipelineName, "found", "secret").
		Return(true, nil)
	mockClient.EXPECT().Exists(man.Team, pipelineName, "not", "found").
		Return(false, nil)
	mockClient.EXPECT().VaultPrefix().Return(man.Repo.GetName())

	result := linter.Lint(man)

	assert.Len(t, result.Errors, 1)
	assert.IsType(t, errors.NotFoundVaultSecretError{}, result.Errors[0])
}
