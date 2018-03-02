package secret_resolver

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
)

type VaultClient interface {
	// This interface just wraps a method in vault.api.logical
	// so we can inject a test client.
	Read(path string) (*api.Secret, error)
}

type SecretResolver interface {
	Exists(path string, secretKey string) (exists bool, err error)
}

type secretResolver struct {
	Fs afero.Afero

	// The following two fields are only used for tests!
	User        user.User   // Afero is not able to expand ~/ so we need to pass in a user object that contains the users HomeDir.
	VaultClient VaultClient // Wrapper interface used to inject a stub
}

func NewSecretResolver(fs afero.Afero) secretResolver {
	return secretResolver{
		Fs: fs,
	}
}

func (secretResolver) getVaultAddr() (err error) {
	vaultHost := os.Getenv("VAULT_ADDR")
	if vaultHost == "" {
		err = NewVaultClientError("Environment variable 'VAULT_ADDR' must be set!")
		return
	}
	return
}

func (s secretResolver) getHomeDir() (homeDir string, error error) {
	// If s.User == user.User{}, i.e a empty UserObject it means we are running production code...
	if (user.User{} == s.User) {
		u, err := user.Current()
		homeDir = u.HomeDir
		error = err
		return
	}

	return s.User.HomeDir, nil
}

func (s secretResolver) readVaultToken() (vaultToken string, err error) {
	homeDir, err := s.getHomeDir()
	if err != nil {
		return
	}

	vaultTokenPath := path.Join(homeDir, ".vault-token")
	content, err := s.Fs.ReadFile(vaultTokenPath)
	if err != nil {
		err = NewVaultClientError(fmt.Sprintf("Could not read vault token from path '%s'", vaultTokenPath))
		return
	}

	_, err = uuid.Parse(string(content))
	if err != nil {
		err = NewVaultClientError(fmt.Sprintf("Content of '%s' does not look like a vault token!", vaultTokenPath))
		return
	}

	vaultToken = string(content)
	return
}

func (s secretResolver) createVaultClient() (vaultClient VaultClient, err error) {
	err = s.getVaultAddr()
	if err != nil {
		return
	}

	vaultToken, err := s.readVaultToken()
	if err != nil {
		return
	}

	if s.VaultClient == nil {
		// If s.VaultClient is nil it means we are running production code
		config := api.DefaultConfig()
		client, e := api.NewClient(config)
		if e != nil {
			err = e
			return
		}

		client.SetToken(vaultToken)
		vaultClient = client.Logical()
		return
	}

	return s.VaultClient, nil
}

func (s secretResolver) Exists(path string, secretKey string) (exists bool, err error) {
	vaultClient, err := s.createVaultClient()
	if err != nil {
		return
	}

	secret, e := vaultClient.Read(path)
	if e != nil {
		err = e
		return
	}

	if secret != nil {
		if _, exists = secret.Data[secretKey]; exists {
			return
		}
	}

	return

}
