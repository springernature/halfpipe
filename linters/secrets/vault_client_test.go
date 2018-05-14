package secrets

import (
	"testing"

	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func createPreClientWithEnvSet() vaultClient {
	preClient := NewVaultClient(afero.Afero{Fs: afero.NewMemMapFs()})
	preClient.getEnv = func(s string) string {
		if s == "VAULT_ADDR" {
			return "https://vault.io"
		}
		return ""
	}
	return preClient
}

func TestVaultAddrNotSet(t *testing.T) {
	preClient := createPreClientWithEnvSet()
	preClient.getEnv = func(s string) string { return "" }

	_, err := preClient.Create()
	assert.Equal(t, errVaultAddrNotSet, err)
}

func TestVaultTokenNotPresent(t *testing.T) {
	preClient := createPreClientWithEnvSet()
	_, err := preClient.Create()

	homeDir, _ := homedir.Dir()
	pathToToken := filepath.Join(homeDir, ".vault-token")

	assert.Equal(t, errVaultTokenMissing(pathToToken), err)
}

func TestVaultTokenInValid(t *testing.T) {
	preClient := createPreClientWithEnvSet()
	homeDir, _ := homedir.Dir()
	pathToToken := filepath.Join(homeDir, ".vault-token")
	preClient.fs.WriteFile(pathToToken, []byte("kehe"), 0777)

	_, err := preClient.Create()

	assert.Equal(t, errVaultTokenInvalid(pathToToken), err)
}

func TestHappyVaultClient(t *testing.T) {
	preClient := createPreClientWithEnvSet()

	homeDir, _ := homedir.Dir()
	preClient.fs.WriteFile(filepath.Join(homeDir, ".vault-token"), []byte("00000000-0000-0000-0000-000000000000"), 0777)

	client, err := preClient.Create()

	assert.Nil(t, err)
	assert.IsType(t, api.Logical{}, *client)
}
