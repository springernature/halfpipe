package secret_resolver

import (
	"os"
	"os/user"
	"path"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var lastPath string
var secret *api.Secret

type MockClient struct{}

func (MockClient) Read(path string) (*api.Secret, error) {
	lastPath = path
	return secret, nil
}

func newSecretResolver() secretResolver {
	lastPath = ""
	secret = nil
	os.Setenv("VAULT_ADDR", "https://vault.io")

	homeDir := "/home/user"
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(path.Join(homeDir, ".vault-token"), []byte("00000000-0000-0000-0000-000000000000"), 0777)

	return secretResolver{
		Fs: fs,
		User: user.User{
			HomeDir: homeDir,
		},
		VaultClient: MockClient{},
	}
}

func assertVaultClientError(t *testing.T, err error) {
	t.Helper()

	_, ok := err.(VaultClientError)
	if !ok {
		assert.Fail(t, "error is not a VaultClientError", err)
	}
}

func TestFailsIfVaultAddrIsNotDefined(t *testing.T) {
	resolver := newSecretResolver()

	os.Unsetenv("VAULT_ADDR")

	_, err := resolver.Exists("", "")

	assert.NotNil(t, err)
	assertVaultClientError(t, err)
	assert.Contains(t, err.Error(), "VAULT_ADDR")
}

func TestFailsIfVaultTokenIsNotPresent(t *testing.T) {
	resolver := newSecretResolver()
	resolver.Fs.Remove(path.Join(resolver.User.HomeDir, ".vault-token"))

	_, err := resolver.Exists("", "")

	assert.NotNil(t, err)
	assertVaultClientError(t, err)
	assert.Contains(t, err.Error(), ".vault-token")
}

func TestFailsIfVaultTokenIsNotAUUID(t *testing.T) {
	resolver := newSecretResolver()
	resolver.Fs.WriteFile(path.Join(resolver.User.HomeDir, ".vault-token"), []byte("kehe"), 0777)

	_, err := resolver.Exists("", "")

	assert.NotNil(t, err)
	assertVaultClientError(t, err)
	assert.Contains(t, err.Error(), ".vault-token")
}

func TestReturnsFalseIfSecretMapNotFound(t *testing.T) {
	secretPath := "/path/to/map"

	resolver := newSecretResolver()

	found, err := resolver.Exists(secretPath, "")

	assert.Nil(t, err)
	assert.Equal(t, secretPath, lastPath)
	assert.False(t, found)
}

func TestReturnsFalseIfSecretMapFoundButSecretKeyNotFound(t *testing.T) {
	secretPath := "/path/to/map"
	secretKey := "yo"

	resolver := newSecretResolver()
	secret = &api.Secret{
		Data: map[string]interface{}{
			"asd": "asd",
		},
	}

	found, err := resolver.Exists(secretPath, secretKey)

	assert.Nil(t, err)
	assert.Equal(t, secretPath, lastPath)
	assert.False(t, found)
}

func TestReturnsTrueIfSecretMapAndSecretKeyFound(t *testing.T) {
	secretPath := "/path/to/map"
	secretKey := "yo"

	resolver := newSecretResolver()
	secret = &api.Secret{
		Data: map[string]interface{}{
			secretKey: "asd",
		},
	}

	found, err := resolver.Exists(secretPath, secretKey)

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, secretPath, lastPath)
}

func TestIntegrationTest(t *testing.T) {
	// This is mostly to test the coverage in Goland..
	// To run this test you need to have logged into vault and be able to read
	// from the path..
	os.Setenv("VAULT_ADDR", "https://vault.halfpipe.io")
	if os.Getenv("HALFPIPE_INTEGRATION_TEST") != "" {
		resolver := NewSecretResolver(afero.Afero{Fs: afero.NewOsFs()})
		found, err := resolver.Exists("/springernature/engineering-enablement/github", "private_key")

		assert.Nil(t, err)
		assert.True(t, found)

		//

		found, err = resolver.Exists("/springernature/engineering-enablement/github", "not_exists")

		assert.Nil(t, err)
		assert.False(t, found)

		//

		found, err = resolver.Exists("/springernature/engineering-enablement/doesnt_exist", "asd")

		assert.Nil(t, err)
		assert.False(t, found)

	}
}
