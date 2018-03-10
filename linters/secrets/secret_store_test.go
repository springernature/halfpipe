package secrets

import (
	"testing"

	"os"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type fakeClient struct {
	Secret *api.Secret
	Error  error
}

func (m *fakeClient) SetSecret(key string, value string) {
	m.Secret = &api.Secret{
		Data: map[string]interface{}{
			key: value,
		},
	}
	m.Error = nil
}

func (m fakeClient) Read(path string) (secret *api.Secret, err error) {
	return m.Secret, m.Error
}

func storeWithFakeClient() (SecretStore, *fakeClient) {
	client := &fakeClient{}
	return secretStore{
		Client: client,
	}, client
}

func TestSecretMapNotFound(t *testing.T) {
	store, _ := storeWithFakeClient()

	found, err := store.Exists("path", "key")

	assert.Nil(t, err)
	assert.False(t, found)
}

func TestSecretKeyNotFound(t *testing.T) {
	store, client := storeWithFakeClient()
	client.SetSecret("somekey", "some value")

	found, err := store.Exists("path", "missingkey")

	assert.Nil(t, err)
	assert.False(t, found)
}

func TestSecretFound(t *testing.T) {
	store, client := storeWithFakeClient()
	client.SetSecret("somekey", "some value")

	found, err := store.Exists("path", "somekey")

	assert.Nil(t, err)
	assert.True(t, found)
}

func TestClientErrorIsPassedBack(t *testing.T) {
	store, client := storeWithFakeClient()
	client.Error = errors.New("blah")

	_, err := store.Exists("path", "somekey")

	assert.Equal(t, client.Error, err)
}

func TestRealClient(t *testing.T) {
	os.Setenv("VAULT_ADDR", "https://vault.io")
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	_, err := NewSecretStore(fs)()
	assert.IsType(t, errVaultTokenMissing, err)

	fs.WriteFile(vaultTokenPath, []byte("00000000-0000-0000-0000-000000000000"), 0777)
	store, err := NewSecretStore(fs)()
	assert.Nil(t, err)
	assert.IsType(t, secretStore{}, store)
}
