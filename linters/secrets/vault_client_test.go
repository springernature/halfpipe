package secrets

import (
	"testing"

	"github.com/hashicorp/vault/api"
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

	assert.Equal(t, errVaultTokenMissing, err)
}

func TestVaultTokenInValid(t *testing.T) {
	preClient := createPreClientWithEnvSet()
	preClient.fs.WriteFile(vaultTokenPath, []byte("kehe"), 0777)

	_, err := preClient.Create()

	assert.Equal(t, errVaultTokenInvalid, err)
}

func TestHappyVaultClient(t *testing.T) {
	preClient := createPreClientWithEnvSet()
	preClient.fs.WriteFile(vaultTokenPath, []byte("00000000-0000-0000-0000-000000000000"), 0777)

	client, err := preClient.Create()

	assert.Nil(t, err)
	assert.IsType(t, api.Logical{}, *client)
}
