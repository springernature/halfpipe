package secrets

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestVaultAddrNotSet(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	_, err := NewVaultClient(afero.Afero{Fs: afero.NewMemMapFs()})
	assert.Equal(t, errVaultAddrNotSet, err)
}

func TestVaultTokenNotPresent(t *testing.T) {
	os.Setenv("VAULT_ADDR", "https://vault.io")
	_, err := NewVaultClient(afero.Afero{Fs: afero.NewMemMapFs()})

	assert.Equal(t, errVaultTokenMissing, err)
}

func TestVaultTokenInValid(t *testing.T) {
	os.Setenv("VAULT_ADDR", "https://vault.io")
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(vaultTokenPath, []byte("kehe"), 0777)
	_, err := NewVaultClient(fs)

	assert.Equal(t, errVaultTokenInvalid, err)
}

func TestHappyVaultClient(t *testing.T) {
	os.Setenv("VAULT_ADDR", "https://vault.io")
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(vaultTokenPath, []byte("00000000-0000-0000-0000-000000000000"), 0777)

	client, err := NewVaultClient(fs)

	assert.Nil(t, err)
	assert.IsType(t, api.Logical{}, *client)
}

// ---------- //

var originalVaultAddr = os.Getenv("VAULT_ADDR")

func TestIntegrationTest(t *testing.T) {
	// This is mostly to test the coverage in Goland..
	// To run this test you need to have logged into vault and be able to read
	// from the path..
	os.Setenv("VAULT_ADDR", originalVaultAddr)
	if os.Getenv("HALFPIPE_INTEGRATION_TEST") != "" {
		store, err := NewSecretStore(afero.Afero{Fs: afero.NewOsFs()})()

		found, err := store.Exists("/springernature/engineering-enablement/github", "private_key")
		assert.Nil(t, err)
		assert.True(t, found)

		found, err = store.Exists("/springernature/engineering-enablement/github", "not_exists")
		assert.Nil(t, err)
		assert.False(t, found)

		found, err = store.Exists("/springernature/engineering-enablement/doesnt_exist", "asd")
		assert.Nil(t, err)
		assert.False(t, found)
	}
}
