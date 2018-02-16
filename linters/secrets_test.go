package linters

import (
	"github.com/springernature/halfpipe/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/golang/mock/gomock"
	"github.com/springernature/halfpipe/vault"
	"fmt"
)

func setupSecretLinter(t *testing.T) (secretLinter SecretsLinter) {
	ctrl := gomock.NewController(t)

	mockClient := vault.NewMockVaultClient(ctrl)
	return SecretsLinter{VaultClient: mockClient}
}

func TestFindSecretsDoesNothingIfThereAreNoSecrets(t *testing.T) {
	man := model.Manifest{}
	result := setupSecretLinter(t).Lint(man)
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

	result := setupSecretLinter(t).Lint(man)
	assert.Len(t, result.Errors, 3)
	assertVaultError(t, "a", result.Errors[0])
	assertVaultError(t, "b", result.Errors[1])
	assertVaultError(t, "c", result.Errors[2])
}

func TestFindSecretsReturnsErrorIfASecretIsMalformed(t *testing.T) {
	key := "((a))"
	man := model.Manifest{
		Repo: model.Repo{
			PrivateKey: key,
		},
	}

	linter := setupSecretLinter(t)
	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertVaultError(t, "a", result.Errors[0])
}

func TestFindSecretsReturnsErrorIfASecretIsNotInVault(t *testing.T) {
	man := model.Manifest{
		Team: "yolo",
		Repo: model.Repo{
			Uri:        "git@github.com:springernature/halfpipe.git",
			PrivateKey: "((deploy.key))",
		},
	}

	path1 := fmt.Sprintf(VaultPathWithRepoName, man.Team, man.Repo.GetName(), "deploy", "key")
	path2 := fmt.Sprintf(VaultPathWithoutRepoName, man.Team, "deploy", "key")

	linter := setupSecretLinter(t)
	ctrl := gomock.NewController(t)
	mockClient := vault.NewMockVaultClient(ctrl)
	mockClient.EXPECT().Exists(path1).Return(false, nil)
	mockClient.EXPECT().Exists(path2).Return(false, nil)
	linter.VaultClient = mockClient

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 1)
	assertVaultError(t, "deploy.key", result.Errors[0])
}

func TestFindSecretsReturnsNoErrorIfASecretIsInVault(t *testing.T) {
	man := model.Manifest{
		Team: "yolo",
		Repo: model.Repo{
			Uri:        "git@github.com:springernature/halfpipe.git",
			PrivateKey: "((deploy.key))",
		},
	}

	path1 := fmt.Sprintf(VaultPathWithRepoName, man.Team, man.Repo.GetName(), "deploy", "key")
	path2 := fmt.Sprintf(VaultPathWithoutRepoName, man.Team, "deploy", "key")

	linter := setupSecretLinter(t)
	ctrl := gomock.NewController(t)
	mockClient := vault.NewMockVaultClient(ctrl)
	mockClient.EXPECT().Exists(path1).Return(false, nil)
	mockClient.EXPECT().Exists(path2).Return(true, nil)
	linter.VaultClient = mockClient

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}

func TestFindSecretsReturnsNoErrorAnsShortCirtcuitsIfASecretIsInVault(t *testing.T) {
	man := model.Manifest{
		Team: "yolo",
		Repo: model.Repo{
			Uri:        "git@github.com:springernature/halfpipe.git",
			PrivateKey: "((deploy.key))",
		},
	}

	path1 := fmt.Sprintf(VaultPathWithRepoName, man.Team, man.Repo.GetName(), "deploy", "key")

	linter := setupSecretLinter(t)
	ctrl := gomock.NewController(t)
	mockClient := vault.NewMockVaultClient(ctrl)
	mockClient.EXPECT().Exists(path1).Return(true, nil)
	linter.VaultClient = mockClient

	result := linter.Lint(man)
	assert.Len(t, result.Errors, 0)
}


// Given this manifest
// I call out in this way
// And if stuff returns from dep
// I give back these errors.
