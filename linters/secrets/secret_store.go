package secrets

import (
	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
)

type SecretReader interface {
	Read(path string) (*api.Secret, error)
}

type SecretStore interface {
	Exists(path string, secretKey string) (exists bool, err error)
}

type secretStore struct {
	Client SecretReader
}

//return a closure so we can defer Newing up until needed
type SecretStoreFunc func() (SecretStore, error)

func NewSecretStore(fs afero.Afero) SecretStoreFunc {
	return func() (store SecretStore, err error) {
		client, err := NewVaultClient(fs)
		if err == nil {
			store = secretStore{Client: client}
		}
		return
	}
}

func (s secretStore) Exists(path string, secretKey string) (exists bool, err error) {
	secret, err := s.Client.Read(path)
	if err != nil || secret == nil {
		return
	}
	_, exists = secret.Data[secretKey]
	return
}
