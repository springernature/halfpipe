package secrets

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/errors"
)

var (
	vaultTokenPath       = filepath.Join(os.Getenv("HOME"), ".vault-token")
	errVaultAddrNotSet   = errors.NewVaultClientErrorf("Environment variable 'VAULT_ADDR' must be set!")
	errVaultTokenMissing = errors.NewVaultClientErrorf("Could not read vault token from path '%s'", vaultTokenPath)
	errVaultTokenInvalid = errors.NewVaultClientErrorf("Contents of '%s' does not look like a vault token!", vaultTokenPath)
)

type vaultClient struct {
	fs     afero.Afero
	getEnv func(string) string
}

func NewVaultClient(fs afero.Afero) vaultClient {
	return vaultClient{
		fs:     fs,
		getEnv: os.Getenv,
	}
}

func (v vaultClient) Create() (vaultClient *api.Logical, err error) {
	if v.getEnv("VAULT_ADDR") == "" {
		err = errVaultAddrNotSet
		return
	}
	vaultTokenPath := vaultTokenPath
	token, err := v.fs.ReadFile(vaultTokenPath)
	if err != nil {
		err = errVaultTokenMissing
		return
	}
	if _, e := uuid.Parse(string(token)); e != nil {
		err = errVaultTokenInvalid
		return
	}
	apiClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return
	}
	apiClient.SetToken(string(token))
	vaultClient = apiClient.Logical()
	return
}
